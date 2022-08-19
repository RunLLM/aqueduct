import { Box } from '@mui/material';
import Link from '@mui/material/Link';
import React, { useEffect, useState } from 'react';

import {
  BigQueryConfig,
  FileData,
  IntegrationConfig,
} from '../../../utils/integrations';
import { IntegrationFileUploadField } from './IntegrationFileUploadField';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: BigQueryConfig = {
  project_id: 'aqueduct_1234',
};

type Props = {
  setDialogConfig: (config: IntegrationConfig) => void;
};

export const BigQueryDialog: React.FC<Props> = ({ setDialogConfig }) => {
  const [projectId, setProjectId] = useState<string>(null);
  const [file, setFile] = useState(null);

  useEffect(() => {
    let contents = null;
    if (file) {
      contents = file.data;
    }
    const config: BigQueryConfig = {
      project_id: projectId,
      service_account_credentials: contents,
    };
    setDialogConfig(config);
  }, [projectId, file]);

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
        spellCheck={false}
        required={true}
        label="Project ID*"
        description="The BigQuery project ID."
        placeholder={Placeholders.project_id}
        onChange={(event) => setProjectId(event.target.value)}
        value={projectId}
      />

      <IntegrationFileUploadField
        label={'Service Account Credentials*'}
        description={fileUploadDescription}
        required={true}
        file={file}
        placeholder={'Upload your service account key file.'}
        onFiles={(files) => {
          const file = files[0];
          readCredentialsFile(file, setFile);
        }}
        displayFile={null}
        onReset={(_) => {
          setFile(null);
        }}
      />
    </Box>
  );
};

export function readCredentialsFile(
  file: File,
  setFile: (credentials: FileData) => void
) {
  const reader = new FileReader();
  reader.onloadend = function (event) {
    const content = event.target.result as string;
    setFile({ name: file.name, data: content });
  };
  reader.readAsText(file);
}
