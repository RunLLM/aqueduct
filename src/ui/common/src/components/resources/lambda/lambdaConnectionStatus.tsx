import Alert from '@mui/material/Alert';
import Box from '@mui/material/Typography';
import Typography from '@mui/material/Typography';
import React from 'react';

import { ExecState, ExecutionStatus } from '../../../utils/shared';
import ExecutionChip from '../../execution/chip';

type Props = {
  state: ExecState;
};

const LambdaConnectionStatus: React.FC<Props> = ({ state }) => {
  const chip = <ExecutionChip status={state.status} />;
  if (state.status === ExecutionStatus.Succeeded) {
    return <Box width="fit-content">{chip}</Box>;
  }

  if (state.status === ExecutionStatus.Failed) {
    return (
      <Box display="flex" flexDirection="column">
        <Box marginBottom={1}>{chip}</Box>
        <Alert severity="error">
          Failed to connect to Lambda resource:{' '}
          <code>{state.error.context}</code>. Once you resolved the error, you
          can delete this resource and retry connection.
        </Alert>
      </Box>
    );
  }

  return (
    <Box display="flex" flexDirection="row" alignItems="center">
      <Box marginRight={1}>{chip}</Box>
      <Typography variant="body2">
        We are still connecting Lambda for you. It usually takes a few minutes
        to complete.
      </Typography>
    </Box>
  );
};

export default LambdaConnectionStatus;
