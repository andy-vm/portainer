import { Edit, Plus } from 'lucide-react';
import { useState } from 'react';
import { FormikErrors } from 'formik';

import { readFileAsText } from '@/portainer/services/fileUploadReact';

import { Button } from '@@/buttons';
import { TextTip } from '@@/Tip/TextTip';
import { FileUploadField } from '@@/form-components/FileUpload';
import { InputList } from '@@/form-components/InputList';
import { ItemProps } from '@@/form-components/InputList/InputList';
import { InputGroup } from '@@/form-components/InputGroup';

import { FormError } from '../FormError';

import { type EnvVar, type Value } from './types';
import { parseDotEnvFile } from './utils';

export function SimpleMode({
  value,
  onChange,
  onAdvancedModeClick,
  errors,
}: {
  value: Value;
  onChange: (value: Value) => void;
  onAdvancedModeClick: () => void;
  errors?: FormikErrors<EnvVar>[];
}) {
  return (
    <>
      <Button
        size="small"
        color="link"
        icon={Edit}
        className="!ml-0 p-0 hover:no-underline"
        onClick={onAdvancedModeClick}
      >
        Advanced mode
      </Button>

      <TextTip color="blue" inline={false}>
        Switch to advanced mode to copy & paste multiple variables
      </TextTip>

      <div className="flex gap-2">
        <Button
          onClick={() => onChange([...value, { name: '', value: '' }])}
          color="default"
          icon={Plus}
        >
          Add an environment variable
        </Button>

        <FileEnv onChooseFile={(add) => onChange([...value, ...add])} />
      </div>

      <InputList
        onChange={onChange}
        value={value}
        isAddButtonHidden
        item={Item}
        errors={errors}
      />
    </>
  );
}

function Item({
  item,
  onChange,
  disabled,
  error,
  readOnly,
}: ItemProps<EnvVar>) {
  return (
    <div className="relative flex w-full flex-col">
      <div className="flex w-full items-center gap-2">
        <InputGroup size="small" className="w-full">
          <InputGroup.Addon>name</InputGroup.Addon>
          <InputGroup.Input
            value={item.name || ''}
            onChange={(e) => handleChange({ name: e.target.value })}
            disabled={disabled}
            readOnly={readOnly}
            placeholder="e.g. FOO"
          />
        </InputGroup>
        <InputGroup size="small" className="w-full">
          <InputGroup.Addon>value</InputGroup.Addon>
          <InputGroup.Input
            value={item.value || ''}
            onChange={(e) => handleChange({ value: e.target.value })}
            disabled={disabled}
            readOnly={readOnly}
            placeholder="e.g. bar"
          />
        </InputGroup>
      </div>

      {!!error && (
        <div className="absolute -bottom-5">
          <FormError className="m-0">{Object.values(error)[0]}</FormError>
        </div>
      )}
    </div>
  );

  function handleChange(partial: Partial<EnvVar>) {
    onChange({ ...item, ...partial });
  }
}

function FileEnv({ onChooseFile }: { onChooseFile: (file: Value) => void }) {
  const [file, setFile] = useState<File | null>(null);

  const fileTooBig = file && file.size > 1024 * 1024;

  return (
    <>
      <FileUploadField
        inputId="env-file-upload"
        onChange={handleChange}
        title="Load variables from .env file"
        accept=".env"
        value={file}
        color="default"
      />

      {fileTooBig && (
        <TextTip color="orange" inline>
          File too large! Try uploading a file smaller than 1MB
        </TextTip>
      )}
    </>
  );

  async function handleChange(file: File) {
    setFile(file);
    if (!file) {
      return;
    }

    const text = await readFileAsText(file);
    if (!text) {
      return;
    }

    const parsed = parseDotEnvFile(text);
    onChooseFile(parsed);
  }
}
