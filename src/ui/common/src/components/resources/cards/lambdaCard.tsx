import Box from '@mui/material/Box';
import React from 'react';

import { Integration, LambdaConfig } from '../../../utils/resources';
import { ExecState, ExecutionStatus } from '../../../utils/shared';
import LambdaConnectionStatus from '../lambda/lambdaConnectionStatus';
import { ResourceCardText, TruncatedText } from './text';

type Props = {
  resource: Integration;
};

export const LambdaCard: React.FC<Props> = ({ resource }) => {
  const config = resource.config as LambdaConfig;
  return (
    <ResourceCardText labels={['Lambda Role ARN']} values={[config.role_arn]} />
  );
};

export const LambdaDetailCard: React.FC<Props> = ({ resource }) => {
  const config = resource.config as LambdaConfig;
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
