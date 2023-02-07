import { object } from 'yup';
import { envVarValidation } from '@@/form-components/EnvironmentVariablesFieldset/EnvironmentVariablesFieldset';
import { validateForm } from '@@/form-components/validate-form';

export default class EnvironmentVariablesPanelController {
  /* @ngInject */
  constructor($async, $scope) {
    this.$async = $async;
    this.$scope = $scope;

    this.handleChange = this.handleChange.bind(this);
  }

  async runValidation(value) {
    return this.$async(async () => {
      this.errors = [];
      if (this.envVarsForm) {
        this.envVarsForm.$setValidity('envVarsForm', true, this.envVarsForm);
      }

      const { value: errors } = await validateField(value);
      this.errors = errors;

      if (Object.keys(this.errors).length > 0) {
        if (this.envVarsForm) {
          this.envVarsForm.$setValidity('envVarsForm', false, this.envVarsForm);
        }
      }
    });
  }

  handleChange(value) {
    this.$scope.$evalAsync(async () => {
      this.onChange(value);
      await this.runValidation(value);
    });
  }

  async $onInit() {
    await this.runValidation(this.value);
  }
}

function validateField(value) {
  return validateForm(() => object({ value: envVarValidation() }), { value });
}
