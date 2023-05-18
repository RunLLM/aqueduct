import { AlertTitle } from '@mui/material';
import { Alert, Box, CircularProgress, Typography } from '@mui/material';
import React from 'react';

import { NodeArtifactResultsGetResponse } from '../../../../handlers/v2/NodeArtifactResultsGet';
import { theme } from '../../../../styles/theme/theme';
import { Data, DataSchema } from '../../../../utils/data';
import { CheckLevel } from '../../../../utils/operators';
import ExecutionStatus, {
  getArtifactResultTableRow,
  stringToExecutionStatus,
} from '../../../../utils/shared';
import { StatusIndicator } from '../../workflowStatus';

type CheckHistoryProps = {
  history: NodeArtifactResultsGetResponse;
  isLoading: boolean;
  error: string;
  checkLevel?: string;
};

const checkHistorySchema: DataSchema = {
  fields: [
    { name: 'status', type: 'varchar' },
    { name: 'level', type: 'varchar' },
    { name: 'value', type: 'varchar' },
    { name: 'timestamp', type: 'varchar' },
  ],
  pandas_version: '', // Not sure what actual value to put here, just filling in for now :)
};

const CheckHistory: React.FC<CheckHistoryProps> = ({
  history,
  isLoading,
  error,
  checkLevel,
}) => {
  if (isLoading) {
    return <CircularProgress />;
  }

  if (error) {
    return (
      <Alert style={{ marginTop: '10px' }} severity="error">
        <AlertTitle>Failed to load historical data.</AlertTitle>
        <pre>{error}</pre>
      </Alert>
    );
  }

  const historicalData: Data = {
    schema: checkHistorySchema,
    data: history.map((result) => {
      const resultRow = getArtifactResultTableRow(result);
      resultRow.level = checkLevel ?? 'undefined';
      return resultRow;
    }),
  };

  const dataSortedByLatest = historicalData.data.sort(
    (x, y) =>
      Date.parse(y['timestamp'] as string) -
      Date.parse(x['timestamp'] as string)
  );

  return (
    <Box mt="32px">
      <Typography variant="h6" fontWeight="normal" marginBottom="8px">
        History
      </Typography>

      {dataSortedByLatest.map((entry, index) => {
        let backgroundColor, hoverColor;
        if (entry.status === ExecutionStatus.Succeeded) {
          backgroundColor = theme.palette.green[100];
          hoverColor = theme.palette.green[200];
        } else if (
          entry.status === ExecutionStatus.Failed &&
          checkLevel === CheckLevel.Error
        ) {
          backgroundColor = theme.palette.red[25];
          hoverColor = theme.palette.red[100];
        } else if (
          entry.status === ExecutionStatus.Failed &&
          checkLevel === CheckLevel.Warning
        ) {
          backgroundColor = theme.palette.yellow[100];
          hoverColor = theme.palette.yellow[200];
        } else {
          // unknown or canceled status
          backgroundColor = theme.palette.gray[100];
          hoverColor = theme.palette.gray[200];
        }

        let checkStatus = entry.status as string;
        if (
          checkLevel === CheckLevel.Warning &&
          entry.status === ExecutionStatus.Failed
        ) {
          checkStatus = ExecutionStatus.Warning as string;
        }

        return (
          <Box
            key={entry.timestamp.toString()}
            p={2}
            sx={{
              display: 'flex',
              alignItems: 'center',
              borderBottom:
                index === historicalData.data.length - 1
                  ? ''
                  : `1px solid ${theme.palette.gray[400]}`,
              backgroundColor: backgroundColor,
              '&:hover': { backgroundColor: hoverColor },
            }}
            width="auto"
          >
            <Box sx={{ display: 'flex', alignItems: 'center' }}>
              <StatusIndicator
                status={stringToExecutionStatus(checkStatus)}
                size={'16px'}
                monochrome={false}
              />

              <Typography sx={{ ml: 1 }} variant="body2">
                {entry.timestamp.toLocaleString()}
              </Typography>
            </Box>
          </Box>
        );
      })}
    </Box>
  );
};

export default CheckHistory;
