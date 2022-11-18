import Box from '@mui/material/Box';
import React from 'react';

import { CondaConfig, Integration } from '../../../utils/integrations';
import { ExecState, ExecutionStatus } from '../../../utils/shared';
import ExecutionChip from '../../execution/chip';

type Props = {
  integration: Integration;
};

export const CondaCard: React.FC<Props> = ({ integration }) => {
  const condaConfig = integration.config as CondaConfig;
  const execState = condaConfig.exec_state
    ? (JSON.parse(condaConfig.exec_state) as ExecState)
    : { status: ExecutionStatus.Unknown };

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <ExecutionChip status={execState.status} />
    </Box>
  );
};
