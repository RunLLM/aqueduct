import { Box, Input, Typography } from '@mui/material';
import Link from '@mui/material/Link';
import React, { useEffect, useState } from 'react';

import { BigQueryConfig, IntegrationConfig } from '../../../utils/integrations';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: BigQueryConfig = {
  project_id: 'aqueduct_1234',
};

type Props = {
  setDialogConfig: (config: IntegrationConfig) => void;
};

export type FileEventTarget = EventTarget & { files: FileList };

export const BigQueryDialog: React.FC<Props> = ({ setDialogConfig }) => {
  const [projectId, setProjectId] = useState<string>(null);
  const [credentials, setCredentials] = useState<string>(null);

  useEffect(() => {
    const config: BigQueryConfig = {
      project_id: projectId,
      service_account_credentials: credentials,
    };
    setDialogConfig(config);
  }, [projectId, credentials]);

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

      <Box sx={{ my: 2 }}>
        <Typography variant="body1" sx={{ fontWeight: 'bold' }}>
          Service Account Credentials*
        </Typography>
        <Typography variant="body2">
          {'Follow the instructions '}
          <Link
            sx={{ fontSize: 'inherit' }}
            target="_blank"
            href="https://cloud.google.com/docs/authentication/getting-started#creating_a_service_account"
          >
            here.
          </Link>
          {' to get your service account key file.'}
        </Typography>

        <Input
          type="file"
          sx={{ marginTop: 2 }}
          placeholder={'Upload your service account key file.'}
          onChange={(event) => {
            const fileEvent: FileEventTarget = event.target as FileEventTarget;
            readCredentialsFile(fileEvent.files[0], setCredentials);
          }}
        />
      </Box>
    </Box>
  );
};

function readCredentialsFile(
  file: File,
  setCredentials: (credentials: string) => void
) {
  const reader = new FileReader();
  reader.onloadend = function (event) {
    const content = event.target.result as string;
    setCredentials(content);
  };
  reader.readAsText(file);
}
