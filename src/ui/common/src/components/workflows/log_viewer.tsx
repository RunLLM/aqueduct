import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import React from 'react';

import { Error, Logs } from '../../utils/shared';
import LogBlock, { LogLevel } from '../text/LogBlock';

type Props = {
  logs?: Logs;
  err?: Error;
};

const LogViewer: React.FC<Props> = ({ logs, err }) => {
  let hasLogs = false;
  if (!!logs && (logs.stdout || logs.stderr)) {
    hasLogs = true;
  }

  if (!hasLogs && !err) {
    return <Alert severity="info">No logs generated.</Alert>;
  }

  return (
    <Box pb={1}>
      {hasLogs && logs.stdout && (
        <Box mb={2}>
          <LogBlock
            logText={logs.stdout}
            title="Logs stdout"
            level={LogLevel.Info}
          />
        </Box>
      )}

      {hasLogs && logs.stderr && (
        <Box mb={2}>
          <LogBlock
            logText={logs.stderr}
            title="Logs stderr"
            level={LogLevel.Info}
          />
        </Box>
      )}

      {!!err && (
        <LogBlock
          logText={`${err.tip ?? ''}\n${err.context ?? ''}`}
          title="Errors"
          level={LogLevel.Error}
        />
      )}
    </Box>
  );
};

export default LogViewer;
