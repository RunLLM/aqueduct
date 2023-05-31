import { Box } from '@mui/material';
import Link from '@mui/material/Link';
import React, { useState } from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import {
  FileData,
  GarConfig,
  ResourceDialogProps,
} from '../../../utils/resources';
import { ResourceFileUploadField } from './ResourceFileUploadField';
import { requiredAtCreate } from './schema';

export const GARDialog: React.FC<ResourceDialogProps<GarConfig>> = () => {
  const { setValue } = useFormContext();
  const [fileData, setFileData] = useState<FileData | null>(null);
  const setFile = (fileData: FileData | null) => {
    setValue('service_account_key', fileData?.data);
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
      <> to get your service account key file.</>
    </>
  );

  return (
    <Box sx={{ mt: 2 }}>
      <ResourceFileUploadField
        name="service_account_key"
        label={'Service Account Key*'}
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

export function getGARValidationSchema(editMode: boolean) {
  return Yup.object().shape({
    service_account_key: requiredAtCreate(
      Yup.string().transform((value) => {
        if (!value?.data) {
          return null;
        }
        return value.data;
      }),
      editMode,
      'Please upload a service account key file'
    ),
  });
}
