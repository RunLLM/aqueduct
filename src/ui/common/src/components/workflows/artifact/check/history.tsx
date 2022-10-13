import { Alert, Box, CircularProgress, Typography } from '@mui/material';
import {
  faCheckCircle,
  faQuestionCircle,
  faXmarkCircle,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import React from 'react';

import { theme } from '../../../../styles/theme/theme';
import ExecutionStatus from '../../../../utils/shared';
import { ArtifactResultsWithLoadingStatus } from '../../../../reducers/artifactResults';
import { Data, DataSchema } from '../../../../utils/data';
import { isFailed, isInitial, isLoading } from '../../../../utils/shared';

type CheckHistoryProps = {
  historyWithLoadingStatus?: ArtifactResultsWithLoadingStatus;
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
  historyWithLoadingStatus,
  checkLevel,
}) => {
  if (
    !historyWithLoadingStatus ||
    isInitial(historyWithLoadingStatus.status) ||
    isLoading(historyWithLoadingStatus.status)
  ) {
    return <CircularProgress />;
  }

  if (isFailed(historyWithLoadingStatus.status)) {
    return (
      <Alert title="Failed to load historical data.">
        {historyWithLoadingStatus.status.err}
      </Alert>
    );
  }

  const historicalData: Data = {
    schema: checkHistorySchema,
    data: (historyWithLoadingStatus.results?.results ?? []).map(
      (artifactStatusResult) => {
        return {
          status: artifactStatusResult.exec_state?.status ?? 'Unknown',
          level: checkLevel ? checkLevel : 'undefined',
          value: artifactStatusResult.content_serialized,
          timestamp: new Date( artifactStatusResult.exec_state?.timestamps?.finished_at).toLocaleString(),
        };
      }
    ),
  };

  return (
    <Box mt="32px">
      <Typography variant="h6" fontWeight="normal">
        History
      </Typography>

      {historicalData.data.map((entry, index) => {
        let backgroundColor, hoverColor, icon;
        if (entry.status === ExecutionStatus.Succeeded) {
          backgroundColor = theme.palette.green[100];
          hoverColor = theme.palette.green[200];
          icon = (
            <FontAwesomeIcon
              icon={faCheckCircle}
              color={theme.palette.green[600]}
            />
          );
        } else if (entry.status === ExecutionStatus.Failed) {
          backgroundColor = theme.palette.red[25];
          hoverColor = theme.palette.red[100];
          icon = (
            <FontAwesomeIcon
              icon={faXmarkCircle}
              color={theme.palette.red[600]}
            />
          );
        } else {
          backgroundColor = theme.palette.gray[100];
          hoverColor = theme.palette.gray[200];
          icon = (
            <FontAwesomeIcon
              icon={faQuestionCircle}
              color={theme.palette.gray[600]}
            />
          );
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
            width="fit-content"
          >
            <Box sx={{ display: 'flex', alignItems: 'center' }} width="fit-content">
              {icon}

              <Typography sx={{ ml: 1 }} variant="body2">
                {entry.timestamp.toLocaleString()}
              </Typography>
            </Box>
          </Box>
        );
      })}
    </Box>
  )
};

export default CheckHistory;
