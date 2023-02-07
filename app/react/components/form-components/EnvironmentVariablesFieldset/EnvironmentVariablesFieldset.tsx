import { useState } from 'react';
import { array, object, SchemaOf, string } from 'yup';
import { FormikErrors } from 'formik';

import { FormSection } from '@@/form-components/FormSection';
import { TextTip } from '@@/Tip/TextTip';

import { AdvancedMode } from './AdvancedMode';
import { SimpleMode } from './SimpleMode';
import { EnvVar, Value } from './types';

export function EnvironmentVariablesFieldset({
  explanation,
  onChange,
  value,
  showHelpMessage,
  errors,
}: {
  explanation?: string;
  value: Value;
  onChange(value: Value): void;
  showHelpMessage: boolean;
  errors: FormikErrors<EnvVar>[];
}) {
  const [simpleMode, setSimpleMode] = useState(true);

  return (
    <FormSection title="Environment variables">
      <div className="form-group">
        {!!explanation && (
          <div className="col-sm-12 environment-variables-panel--explanation">
            {explanation}
          </div>
        )}

        <div className="col-sm-12">
          {simpleMode ? (
            <SimpleMode
              onAdvancedModeClick={() => setSimpleMode(false)}
              onChange={onChange}
              value={value}
              errors={errors}
            />
          ) : (
            <AdvancedMode
              onSimpleModeClick={() => setSimpleMode(true)}
              onChange={onChange}
              value={value}
            />
          )}
        </div>

        {showHelpMessage && (
          <TextTip color="blue" inline={false}>
            Environment changes will not take effect until redeployment occurs
            manually or via webhook.
          </TextTip>
        )}
      </div>
    </FormSection>
  );
}

export function envVarValidation(): SchemaOf<Value> {
  return array(
    object({
      name: string().required('Name is required'),
      value: string().default(''),
    })
  );
}
