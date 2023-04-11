import { Box } from '@mui/material';
import Link from '@mui/material/Link';
import React, { useState } from 'react';

import { BigQueryConfig, FileData } from '../../../utils/integrations';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationFileUploadField } from './IntegrationFileUploadField';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: BigQueryConfig = {
  project_id: 'aqueduct_1234',
};

type Props = {
  onUpdateField: (field: keyof BigQueryConfig, value: string) => void;
  value?: BigQueryConfig;
  editMode: boolean;
};

export const BigQueryDialog: React.FC<Props> = ({
  onUpdateField,
  value,
  editMode,
}) => {
  const [fileName, setFileName] = useState<string>(null);
  const setFile = (fileData: FileData | null) => {
    setFileName(fileData?.name ?? null);
    onUpdateField('service_account_credentials', fileData?.data);
  };

  const fileData =
    fileName && !!value?.service_account_credentials
      ? {
        name: fileName,
        data: value.service_account_credentials,
      }
      : null;

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
        onChange={(event) => onUpdateField('project_id', event.target.value)}
        value={value?.project_id ?? null}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationFileUploadField
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
