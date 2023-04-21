import Box from '@mui/material/Box';
import React from 'react';

import { Integration, LambdaConfig } from '../../../utils/integrations';
import { ExecState, ExecutionStatus } from '../../../utils/shared';
import LambdaConnectionStatus from '../lambda/lambdaConnectionStatus';
import { CardTextEntry, TruncatedText } from './text';

type Props = {
  integration: Integration;
};

// This should be set to the minimum width required to display the longest category name on the card.
const categoryWidth = '120px';

export const LambdaCard: React.FC<Props> = ({ integration }) => {
  const config = integration.config as LambdaConfig;
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <CardTextEntry
        category="Lambda Role ARN: "
        value={config.role_arn}
        categoryWidth={categoryWidth}
      />
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
