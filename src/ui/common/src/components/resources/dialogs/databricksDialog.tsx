import { Link } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';
import { useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

import {
  DatabricksConfig,
  IntegrationDialogProps,
} from '../../../utils/resources';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: DatabricksConfig = {
  workspace_url: 'https://dbc-your-workspace.cloud.databricks.com',
  access_token: 'dapi123456789',
  s3_instance_profile_arn:
    'arn:aws:iam::123:instance-profile/access-databuckets-arn',
  instance_pool_id: '123-456-789',
};

export const DatabricksDialog: React.FC<IntegrationDialogProps> = ({
  editMode = false,
}) => {
  const { setValue } = useFormContext();

  return (
    <Box sx={{ mt: 2 }}>
      <Typography variant="body2">
        For more details on connecting to Databricks, please refer{' '}
        <Link href="https://docs.aqueducthq.com/resources/compute-systems/databricks">
          the Aqueduct documentation
        </Link>
        .
      </Typography>
      <IntegrationTextInputField
        name="workspace_url"
        label={'Workspace URL*'}
        description={'URL of Databricks Workspace.'}
        spellCheck={false}
        required={true}
        placeholder={Placeholders.workspace_url}
        onChange={(event) => setValue('workspace_url', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <IntegrationTextInputField
        name="access_token"
        label={'Access Token*'}
        description={
          'The access token to connect to your Databricks Workspace.'
        }
        spellCheck={false}
        required={true}
        placeholder={Placeholders.access_token}
        onChange={(event) => setValue('access_token', event.target.value)}
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <Typography variant="body2">
        For more details on creating an S3 profile for Databricks, please see{' '}
        <Link href="https://docs.databricks.com/aws/iam/instance-profile-tutorial.html">
          the Databricks documentation
        </Link>
        .
      </Typography>

      <IntegrationTextInputField
        name="s3_instance_profile_arn"
        label={'S3 Instance Profile ARN*'}
        description={
          'The ARN of the instance profile that allows Databricks clusters to access S3.'
        }
        spellCheck={false}
        required={true}
        placeholder={Placeholders.s3_instance_profile_arn}
        onChange={(event) =>
          setValue('s3_instance_profile_arn', event.target.value)
        }
        disabled={editMode}
        warning={editMode ? undefined : readOnlyFieldWarning}
        disableReason={editMode ? readOnlyFieldDisableReason : undefined}
      />

      <Typography variant="body2">
        For more details on Databricks Instance Pools, please see{' '}
        <Link href="https://docs.databricks.com/aws/iam/instance-profile-tutorial.html">
          the Databricks documentation
        </Link>
        .
      </Typography>

      <IntegrationTextInputField
        name="instance_pool_id"
        label={'Instance Pool ID'}
        description={
          'The ID of the Databricks Instance Pool that Aqueduct will run compute on.'
        }
        spellCheck={false}
        required={false}
        placeholder={Placeholders.instance_pool_id}
        onChange={(event) => setValue('instance_pool_id', event.target.value)}
      />
    </Box>
  );
};

export function getDatabricksValidationSchema() {
  return Yup.object().shape({
    workspace_url: Yup.string().required('Please enter a workspace URL'),
    access_token: Yup.string().required('Please enter an access token'),
    s3_instance_profile_arn: Yup.string().required(
      'Please enter an instance profile ARN'
    ),
    instance_pool_id: Yup.string(),
  });
}
