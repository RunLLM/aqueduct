import Box from '@mui/material/Box';
import React from 'react';

import { Integration, LambdaConfig } from '../../../utils/integrations';
import { ExecState, ExecutionStatus } from '../../../utils/shared';
import LambdaConnectionStatus from '../lambda/lambdaConnectionStatus';
import { TruncatedText } from './text';

type Props = {
  integration: Integration;
};

export const LambdaCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as LambdaConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <TruncatedText variant="body2">
        <strong>Lambda Role ARN: </strong>
        {config.role_arn}
      </TruncatedText>
    </Box>
  );
};

export const LambdaDetailCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as LambdaConfig;
  const execState = config.exec_state
    ? (JSON.parse(config.exec_state) as ExecState)
    : { status: ExecutionStatus.Unknown };
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <TruncatedText variant="body2">
        <strong>Lambda Role ARN: </strong>
        {config.role_arn}
      </TruncatedText>
      <LambdaConnectionStatus state={execState} />
    </Box>
  );
};
