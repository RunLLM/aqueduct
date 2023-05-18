import { Box } from '@mui/material';
import Link from '@mui/material/Link';
import React, { useState } from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import {
  BigQueryConfig,
  FileData,
  ResourceDialogProps,
} from '../../../utils/resources';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { ResourceFileUploadField } from './ResourceFileUploadField';
import { ResourceTextInputField } from './ResourceTextInputField';

const Placeholders: BigQueryConfig = {
  project_id: 'aqueduct_1234',
};

export const BigQueryDialog: React.FC<ResourceDialogProps> = ({
  editMode = false,
}) => {
  const { setValue } = useFormContext();
  const [fileData, setFileData] = useState<FileData | null>(null);
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
        href="https://cloud.google.com/docs/authentication/getting-started#creating_a_service_account"
      >
        here
      </Link>
      <> to get your service account key file.</>
    </>
  );

  return (
    <Box sx={{ mt: 2 }}>
      <ResourceTextInputField
        name="project_id"
        spellCheck={false}
        required={true}
        label="Project ID*"
        description="The BigQuery project ID."
        placeholder={Placeholders.project_id}
        onChange={(event) => setValue('project_id', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
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

export function getBigQueryValidationSchema() {
  return Yup.object().shape({
    name: Yup.string().required('Please enter a name'),
    project_id: Yup.string().required('Please enter a project ID'),
    service_account_credentials: Yup.string()
      .transform((value) => {
        if (!value?.data) {
          return null;
        }
        return value.data;
      })
      .required('Please upload a service account key file'),
  });
}
