package edgestacks

import (
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	"github.com/portainer/portainer/api/internal/edge"
	"github.com/portainer/portainer/api/internal/set"
	"github.com/rs/zerolog/log"
)

type updateEdgeStackPayload struct {
	StackFileContent string
	UpdateVersion    bool
	EdgeGroups       []portainer.EdgeGroupID
	DeploymentType   portainer.EdgeStackDeploymentType
	// Uses the manifest's namespaces instead of the default one
	UseManifestNamespaces bool
}

func (payload *updateEdgeStackPayload) Validate(r *http.Request) error {
	if payload.StackFileContent == "" {
		return errors.New("Invalid stack file content")
	}
	if len(payload.EdgeGroups) == 0 {
		return errors.New("Edge Groups are mandatory for an Edge stack")
	}
	return nil
}

// @id EdgeStackUpdate
// @summary Update an EdgeStack
// @description **Access policy**: administrator
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path string true "EdgeStack Id"
// @param body body updateEdgeStackPayload true "EdgeStack data"
// @success 200 {object} portainer.EdgeStack
// @failure 500
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks/{id} [put]
func (handler *Handler) edgeStackUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid stack identifier route variable", err)
	}

	stack, err := handler.DataStore.EdgeStack().EdgeStack(portainer.EdgeStackID(stackID))
	if err != nil {
		return handler.handlerDBErr(err, "Unable to find a stack with the specified identifier inside the database")
	}

	var payload updateEdgeStackPayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	relationConfig, err := edge.FetchEndpointRelationsConfig(handler.DataStore)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve environments relations config from database", err)
	}

	relatedEndpointIds, err := edge.EdgeStackRelatedEndpoints(stack.EdgeGroups, relationConfig.Endpoints, relationConfig.EndpointGroups, relationConfig.EdgeGroups)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve edge stack related environments from database", err)
	}

	groupsIds := stack.EdgeGroups
	if payload.EdgeGroups != nil {
		newRelated, _, err := handler.handleChangeEdgeGroups(stack.ID, payload.EdgeGroups, relatedEndpointIds, relationConfig)
		if err != nil {
			return httperror.InternalServerError("Unable to handle edge groups change", err)
		}

		groupsIds = payload.EdgeGroups
		relatedEndpointIds = newRelated

	}

	entryPoint := stack.EntryPoint
	manifestPath := stack.ManifestPath
	deploymentType := stack.DeploymentType

	if deploymentType != payload.DeploymentType {
		// deployment type was changed - need to delete the old file
		err = handler.FileService.RemoveDirectory(stack.ProjectPath)
		if err != nil {
			log.Warn().Err(err).Msg("Unable to clear old files")
		}

		entryPoint = ""
		manifestPath = ""
		deploymentType = payload.DeploymentType
	}

	stackFolder := strconv.Itoa(int(stack.ID))

	if deploymentType == portainer.EdgeStackDeploymentCompose {
		if entryPoint == "" {
			entryPoint = filesystem.ComposeFileDefaultName
		}

		_, err := handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, entryPoint, []byte(payload.StackFileContent))
		if err != nil {
			return httperror.InternalServerError("Unable to persist updated Compose file on disk", err)
		}

		tempManifestPath, err := handler.convertAndStoreKubeManifestIfNeeded(stackFolder, stack.ProjectPath, entryPoint, relatedEndpointIds)
		if err != nil {
			return httperror.InternalServerError("Unable to convert and persist updated Kubernetes manifest file on disk", err)
		}

		manifestPath = tempManifestPath
	}

	if deploymentType == portainer.EdgeStackDeploymentKubernetes {
		if manifestPath == "" {
			manifestPath = filesystem.ManifestFileDefaultName
		}

		hasDockerEndpoint, err := hasDockerEndpoint(handler.DataStore.Endpoint(), relatedEndpointIds)
		if err != nil {
			return httperror.InternalServerError("Unable to check for existence of docker environment", err)
		}

		if hasDockerEndpoint {
			return httperror.BadRequest("Edge stack with docker environment cannot be deployed with kubernetes config", err)
		}

		_, err = handler.FileService.StoreEdgeStackFileFromBytes(stackFolder, manifestPath, []byte(payload.StackFileContent))
		if err != nil {
			return httperror.InternalServerError("Unable to persist updated Kubernetes manifest file on disk", err)
		}
	}

	err = handler.DataStore.EdgeStack().UpdateEdgeStackFunc(stack.ID, func(edgeStack *portainer.EdgeStack) {
		edgeStack.NumDeployments = len(relatedEndpointIds)
		if payload.UpdateVersion {
			edgeStack.Status = make(map[portainer.EndpointID]portainer.EdgeStackStatus)
			edgeStack.Version++
		}

		edgeStack.UseManifestNamespaces = payload.UseManifestNamespaces

		edgeStack.DeploymentType = deploymentType
		edgeStack.EntryPoint = entryPoint
		edgeStack.ManifestPath = manifestPath

		edgeStack.EdgeGroups = groupsIds
	})
	if err != nil {
		return httperror.InternalServerError("Unable to persist the stack changes inside the database", err)
	}

	return response.JSON(w, stack)
}

func (handler *Handler) handleChangeEdgeGroups(edgeStackID portainer.EdgeStackID, newEdgeGroupsIDs []portainer.EdgeGroupID, oldRelatedEnvironmentIDs []portainer.EndpointID, relationConfig *edge.EndpointRelationsConfig) ([]portainer.EndpointID, set.Set[portainer.EndpointID], error) {
	newRelatedEnvironmentIDs, err := edge.EdgeStackRelatedEndpoints(newEdgeGroupsIDs, relationConfig.Endpoints, relationConfig.EndpointGroups, relationConfig.EdgeGroups)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "Unable to retrieve edge stack related environments from database")
	}

	oldRelatedSet := set.ToSet(oldRelatedEnvironmentIDs)
	newRelatedSet := set.ToSet(newRelatedEnvironmentIDs)

	endpointsToRemove := set.Set[portainer.EndpointID]{}
	for endpointID := range oldRelatedSet {
		if !newRelatedSet[endpointID] {
			endpointsToRemove[endpointID] = true
		}
	}

	for endpointID := range endpointsToRemove {
		relation, err := handler.DataStore.EndpointRelation().EndpointRelation(endpointID)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "Unable to find environment relation in database")
		}

		delete(relation.EdgeStacks, edgeStackID)

		err = handler.DataStore.EndpointRelation().UpdateEndpointRelation(endpointID, relation)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "Unable to persist environment relation in database")
		}
	}

	endpointsToAdd := set.Set[portainer.EndpointID]{}
	for endpointID := range newRelatedSet {
		if !oldRelatedSet[endpointID] {
			endpointsToAdd[endpointID] = true
		}
	}

	for endpointID := range endpointsToAdd {
		relation, err := handler.DataStore.EndpointRelation().EndpointRelation(endpointID)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "Unable to find environment relation in database")
		}

		relation.EdgeStacks[edgeStackID] = true

		err = handler.DataStore.EndpointRelation().UpdateEndpointRelation(endpointID, relation)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "Unable to persist environment relation in database")
		}
	}

	return newRelatedEnvironmentIDs, endpointsToAdd, nil
}
