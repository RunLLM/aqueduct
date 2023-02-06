import Box from '@mui/material/Box';
import React from 'react';

import { DatabricksConfig } from '../../../utils/integrations';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: DatabricksConfig = {
  workspace_url: 'workspace_url',
  access_token: 'access_token',
  s3_instance_profile_arn: 's3_instance_profile_arn',
};

type Props = {
  onUpdateField: (field: keyof DatabricksConfig, value: string) => void;
  value?: DatabricksConfig;
  editMode: boolean;
};

export const DatabricksDialog: React.FC<Props> = ({
  onUpdateField,
  value,
  editMode,
}) => {
  return (
    <Box sx={{ mt: 2 }}>
      <IntegrationTextInputField
        label={'Workspace URL*'}
        description={'URL of Databricks Workspace.'}
        spellCheck={false}
        required={true}
        placeholder={Placeholders.workspace_url}
        onChange={(event) => onUpdateField('workspace_url', event.target.value)}
        value={value?.workspace_url ?? null}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        label={'Access Token*'}
        description={
          'The access token to connect to your Databricks Workspace.'
        }
        spellCheck={false}
        required={true}
        placeholder={Placeholders.access_token}
        onChange={(event) => onUpdateField('access_token', event.target.value)}
        value={value?.access_token ?? null}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        label={'S3 Instance Profile ARN*'}
        description={
          'The ARN of the instance profile that allows Databricks clusters to access S3.'
        }
        spellCheck={false}
        required={true}
        placeholder={Placeholders.s3_instance_profile_arn}
        onChange={(event) =>
          onUpdateField('s3_instance_profile_arn', event.target.value)
        }
        value={value?.s3_instance_profile_arn ?? null}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />
    </Box>
  );
};

export function isDatabricksConfigComplete(config: DatabricksConfig): boolean {
  return (
    !!config.access_token &&
    !!config.s3_instance_profile_arn &&
    !!config.workspace_url
  );
}
