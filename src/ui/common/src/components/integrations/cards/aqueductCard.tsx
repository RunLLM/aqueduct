import { Box, Typography } from '@mui/material';
import React from 'react';

import {
  AqueductComputeConfig,
  CondaConfig,
  Integration,
} from '../../../utils/integrations';
import { ExecState, ExecutionStatus } from '../../../utils/shared';
import { ResourceCardText } from './text';

type Props = {
  integration: Integration;
  detailedView: boolean;
};

export const AqueductCard: React.FC<Props> = ({
  integration,
  detailedView,
}) => {
  const config = integration.config as AqueductComputeConfig;

  if (config.python_version) {
    const tokenized_python_version = config.python_version.split(' ');
    if (tokenized_python_version.length != 2) {
      return null;
    }
    return (
      <ResourceCardText
        labels={['Python Version']}
        values={[tokenized_python_version[1]]}
      />
    );
  } else if (config.conda_config_serialized) {
    const conda_config = JSON.parse(
      config.conda_config_serialized
    ) as CondaConfig;

    // Only use ResourceCardText in the detailed view.
    if (detailedView) {
      return (
        <ResourceCardText
          labels={['Conda Path']}
          values={[conda_config.conda_path]}
        />
      );
    }

    // For an Aqueduct + Conda summary card, display a message about Conda's current status.
    const conda_exec_state = JSON.parse(conda_config.exec_state) as ExecState;
    const finished_at: string | undefined =
      conda_exec_state.timestamps?.finished_at;

    let conda_msg: string;
    if (conda_exec_state.status === ExecutionStatus.Succeeded && finished_at) {
      conda_msg =
        'Conda connected on ' + new Date(finished_at).toLocaleString();
    } else if (
      conda_exec_state.status == ExecutionStatus.Failed &&
      finished_at
    ) {
      conda_msg =
        'Conda failed to connect on ' + new Date(finished_at).toLocaleString();
    } else {
      conda_msg = 'Conda connecting...';
    }

    return (
      <Box sx={{ display: 'flex', flexDirection: 'row', alignItems: 'center' }}>
        <Typography variant="caption" sx={{ fontWeight: 300 }}>
          {conda_msg}
        </Typography>
      </Box>
    );
  }
  return null;
};
