import {
  faCheckCircle,
  faQuestionCircle,
  faXmarkCircle,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { CircularProgress, Typography } from '@mui/material';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import React from 'react';
import Plot from 'react-plotly.js';

import { ArtifactResultsWithLoadingStatus } from '../../../../reducers/artifactResults';
import { theme } from '../../../../styles/theme/theme';
import { Data, DataSchema } from '../../../../utils/data';
import ExecutionStatus, {
  isFailed,
  isInitial,
  isLoading,
} from '../../../../utils/shared';

type Props = {
  historyWithLoadingStatus?: ArtifactResultsWithLoadingStatus;
};

const metricHistorySchema: DataSchema = {
  fields: [
    { name: 'status', type: 'varchar' },
    { name: 'timestamp', type: 'varchar' },
    { name: 'value', type: 'float' },
  ],
  pandas_version: '',
};

const MetricsHistory: React.FC<Props> = ({ historyWithLoadingStatus }) => {
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
    schema: metricHistorySchema,
    data: (historyWithLoadingStatus.results?.results ?? []).map(
      (artifactStatusResult) => {
        return {
          status: artifactStatusResult.exec_state?.status ?? 'Unknown',
          timestamp: new Date(
            artifactStatusResult.exec_state?.timestamps?.finished_at
          ).toLocaleString(),
          value: artifactStatusResult.content_serialized,
        };
      }
    ),
  };
  const dataSortedByLatest = historicalData.data.sort(
    (x, y) =>
      Date.parse(y['timestamp'] as string) -
      Date.parse(x['timestamp'] as string)
  );
  const dataToPlot = historicalData.data
    .filter((x) => !!x['timestamp'] && !!x['value'])
    .sort(
      (x, y) =>
        Date.parse(x['timestamp'] as string) -
        Date.parse(y['timestamp'] as string)
    );
  const timestamps = dataToPlot.map((x) => x['timestamp']);
  const values = dataToPlot.map((x) => x['value']);

  return (
    <Box display="flex" justifyContent="center" flexDirection="column">
      <Box mb={2}>
        <Typography
          variant="h6"
          component="div"
          marginBottom="8px"
          fontWeight="normal"
        >
          History
        </Typography>

        <Plot
          data={[
            {
              x: timestamps,
              y: values,
              type: 'scatter',
              mode: 'lines+markers',
              marker: { color: theme.palette.blue[900] },
              line: { color: theme.palette.blue[900] },
            },
          ]}
          layout={{
            width: '100%',
            height: '100%',
            plot_bgcolor: theme.palette.gray[100],
            margin: { b: 30, l: 50, t: 0, r: 0 },
            yaxis: { ticksuffix: ' ' },
          }}
        />
      </Box>

      <Box width="100%" mt="32px">
        <Typography variant="h6" fontWeight="normal">
          History
        </Typography>
        {dataSortedByLatest.map((entry, index) => {
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
            >
              <Box sx={{ flex: 1, display: 'flex', alignItems: 'center' }}>
                {icon}

                <Typography sx={{ ml: 1 }} variant="body2">
                  {entry.timestamp.toLocaleString()}
                </Typography>
              </Box>
              <Typography variant="body1">{entry.value.toString()}</Typography>
            </Box>
          );
        })}
      </Box>
    </Box>
  );
};

export default MetricsHistory;
