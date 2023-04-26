import { Box } from '@mui/material';
import Link from '@mui/material/Link';
import React, { useState } from 'react';
import { useFormContext } from 'react-hook-form';

import {
  BigQueryConfig,
  FileData,
  IntegrationDialogProps,
} from '../../../utils/integrations';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationFileUploadField } from './IntegrationFileUploadField';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: BigQueryConfig = {
  project_id: 'aqueduct_1234',
};

export const BigQueryDialog: React.FC<IntegrationDialogProps> = ({
  editMode = false,
}) => {
  const [fileData, setFileData] = useState<FileData | null>(null);

  const { setValue } = useFormContext();

  const setFile = (fileData: FileData | null) => {
    // Update the react-hook-form value.
    setValue('service_account_credentials', fileData?.data);
    // Set state to trigger re-render of file upload field.
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
      <IntegrationTextInputField
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

      <IntegrationFileUploadField
        name="service_account_credentials"
        label={'Service Account Credentials*'}
        description={fileUploadDescription}
        required={true}
        file={fileData}
        placeholder={'Upload your service account key file.'}
        onFiles={(files: FileList) => {
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

export function isBigQueryDialogConfigComplete(
  config: BigQueryConfig
): boolean {
  return !!config.project_id && !!config.service_account_credentials;
}
