import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import React from 'react';

import LogBlock, { LogLevel } from '../text/LogBlock';

type Props = {
  logs: { [name: string]: string };
  err: string;
};

const LogViewer: React.FC<Props> = ({ logs, err }) => {
  let hasLogs = false;
  Object.keys(logs).forEach((logKey) => {
    if (logs[logKey].length > 0) {
      hasLogs = true;
    }
  });

  if (!hasLogs && !err) {
    return <Alert severity="info">No logs generated.</Alert>;
  }

  return (
    <Box pb={1}>
      {hasLogs && (
        <Box mb={2}>
          <LogBlock logText={logs.stdout} title="Logs" level={LogLevel.Info} />
        </Box>
      )}

      {err && <LogBlock logText={err} title="Errors" level={LogLevel.Error} />}
    </Box>
  );
};

export default LogViewer;
