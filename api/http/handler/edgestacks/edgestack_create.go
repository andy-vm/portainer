package edgestacks

import (
	"net/http"

	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	httperrors "github.com/portainer/portainer/api/http/errors"
	"github.com/portainer/portainer/api/http/security"
)

// @id EdgeStackCreate
// @summary Create an EdgeStack
// @description **Access policy**: administrator
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param method query string true "Creation Method" Enums(file,string,repository)
// @param body_string body edgeStackCreateTextPayload true "Required when using method=string"
// @param body_file body edgeStackCreateFilePayload true "Required when using method=file"
// @param body_repository body edgeStackCreateGitRepositoryPayload true "Required when using method=repository"
// @success 200 {object} portainer.EdgeStack
// @failure 500
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks [post]
func (handler *Handler) edgeStackCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	method, err := request.RetrieveQueryParameter(r, "method", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: method", err)
	}
	dryrun, _ := request.RetrieveBooleanQueryParameter(r, "dryrun", true)

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user details from authentication token", err)
	}

	edgeStack, err := handler.createSwarmStack(method, dryrun, tokenData.ID, r)
	if err != nil {
		var payloadError *httperrors.InvalidPayloadError
		switch {
		case errors.As(err, &payloadError):
			return httperror.BadRequest("Invalid payload", err)
		case errors.Is(err, &httperrors.ConflictError{}):
			return httperror.NewError(http.StatusConflict, err.Error(), err)
		default:
			return httperror.InternalServerError("Unable to create Edge stack", err)
		}
	}

	return response.JSON(w, edgeStack)
}

func (handler *Handler) createSwarmStack(method string, dryrun bool, userID portainer.UserID, r *http.Request) (*portainer.EdgeStack, error) {

	switch method {
	case "string":
		return handler.createSwarmStackFromFileContent(r, dryrun)
	case "repository":
		return handler.createSwarmStackFromGitRepository(r, dryrun, userID)
	case "file":
		return handler.createSwarmStackFromFileUpload(r, dryrun)
	}
	return nil, errors.New("Invalid value for query parameter: method. Value must be one of: string, repository or file")
}
