import { Checkbox, FormControlLabel } from '@mui/material';
import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
import { useController, useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import {
  FileData,
  GCSConfig,
  ResourceDialogProps,
} from '../../../utils/resources';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { ResourceFileUploadField } from './ResourceFileUploadField';
import { ResourceTextInputField } from './ResourceTextInputField';
import { requiredAtCreate } from './schema';

const Placeholders: GCSConfig = {
  bucket: 'aqueduct',
  use_as_storage: '',
};

interface GCSDialogProps extends ResourceDialogProps<GCSConfig> {
  setMigrateStorage: React.Dispatch<React.SetStateAction<boolean>>;
}

export const GCSDialog: React.FC<GCSDialogProps> = ({
  resourceToEdit,
  setMigrateStorage,
}) => {
  // Setup for the checkbox component.
  const { control, register, setValue } = useFormContext();
  const editMode = !!resourceToEdit;
  if (resourceToEdit) {
    Object.entries(resourceToEdit).forEach(([k, v]) => {
      register(k, { value: v });
    });
  }

  const initialUseAsStorage = resourceToEdit?.use_as_storage ?? 'false';

  const [fileData, setFileData] = useState<FileData | null>(null);
  const { field } = useController({
    control,
    name: 'use_as_storage',
    defaultValue: initialUseAsStorage,
    rules: { required: true },
  });

  const setFile = (fileData: FileData | null) => {
    setValue('service_account_credentials', fileData?.data);
    setFileData(fileData);
  };

  const fileUploadDescription = (
    <>
      <>Follow the instructions </>
      <Link
        sx={{ fontSize: 'inherit' }}
        target="_blank"
        href="https://cloud.google.com/iam/docs/service-accounts-create"
      >
        here
      </Link>
      <> to create a service account and get the service account key file.</>
    </>
  );

  useEffect(() => {
    if (setMigrateStorage) {
      setMigrateStorage(initialUseAsStorage === 'true');
    }
  }, [setMigrateStorage, initialUseAsStorage]);

  return (
    <Box sx={{ mt: 2 }}>
      <ResourceTextInputField
        name="bucket"
        spellCheck={false}
        required={true}
        label="Bucket*"
        description="The name of the GCS bucket."
        placeholder={Placeholders.bucket}
        onChange={(event) => setValue('bucket', event.target.value)}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disabled={editMode}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <ResourceFileUploadField
        name="service_account_credentials"
        label={'Service Account Credentials*'}
        description={fileUploadDescription}
        required={true}
        file={fileData}
        placeholder={'Upload your service account key file.'}
        onFiles={(files) => {
          const file = files[0];
          readCredentialsFile(file, setFile);
        }}
        displayFile={null}
        onReset={() => {
          setFile(null);
        }}
      />

      <FormControlLabel
        label="Use this resource for Aqueduct metadata storage."
        control={
          <Checkbox
            ref={field.ref}
            checked={field.value === 'true'}
            onChange={(event) => {
              const updatedValue = event.target.checked ? 'true' : 'false';
              field.onChange(updatedValue);
            }}
            disabled={true}
          />
        }
      />

      <Typography>
        We currently only support using Google Cloud Storage as the Aqueduct
        metadata storage. Support for using it as a data resource will be added
        soon.
      </Typography>
    </Box>
  );
};

export function readCredentialsFile(
  file: File,
  setFile: (credentials: FileData) => void
): void {
  const reader = new FileReader();
  reader.onloadend = function (event) {
    const content = event.target.result as string;
    setFile({ name: file.name, data: content });
  };
  reader.readAsText(file);
}

export function getGCSValidationSchema(editMode: boolean) {
  return Yup.object().shape({
    name: Yup.string().required('Please enter a name'),
    bucket: Yup.string().required('Please enter a bucket name'),
    service_account_credentials: requiredAtCreate(
      Yup.string().transform((value) => {
        if (!value?.data) {
          return null;
        }

        return value.data;
      }),
      editMode,
      'Please upload a service account key file.'
    ),
  });
}
