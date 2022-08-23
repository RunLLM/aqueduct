import { Checkbox, FormControlLabel } from '@mui/material';
import Box from '@mui/material/Box';
import React, { useEffect, useState } from 'react';

import { GCSConfig, IntegrationConfig } from '../../../utils/integrations';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: GCSConfig = {
  bucket: 'aqueduct',
  credentials_path: '',
  use_as_storage: '',
};

type Props = {
  setDialogConfig: (config: IntegrationConfig) => void;
};

export const GCSDialog: React.FC<Props> = ({ setDialogConfig }) => {
  const [bucket, setBucket] = useState<string>(null);
  const [credentialsPath, setCredentialsPath] = useState<string>(null);
  const [useAsStorage, setUseAsStorage] = useState<string>('false');

  useEffect(() => {
    const config: GCSConfig = {
      bucket: bucket,
      credentials_path: credentialsPath,
      use_as_storage: useAsStorage,
    };
    setDialogConfig(config);
  }, [bucket, credentialsPath, useAsStorage]);

  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Bucket*"
        description="The name of the GCS bucket."
        placeholder={Placeholders.bucket}
        onChange={(event) => setBucket(event.target.value)}
        value={bucket}
      />

      <IntegrationTextInputField
        spellCheck={false}
        required={true}
        label="Credentials Path*"
        description="The filepath to the service account credentials key."
        placeholder={Placeholders.credentials_path}
        onChange={(event) => setCredentialsPath(event.target.value)}
        value={credentialsPath}
      />

      <FormControlLabel
        label="Use this integration for Aqueduct metadata storage."
        control={
          <Checkbox
            checked={useAsStorage === 'true'}
            onChange={(event) =>
              setUseAsStorage(event.target.checked ? 'true' : 'false')
            }
          />
        }
      />
    </Box>
  );
};
