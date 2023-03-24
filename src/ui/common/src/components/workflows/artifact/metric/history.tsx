import { AlertTitle } from '@mui/material';
import { CircularProgress, Typography } from '@mui/material';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import React from 'react';
import Plot from 'react-plotly.js';

import { ArtifactResultsWithLoadingStatus } from '../../../../reducers/artifactResults';
import { theme } from '../../../../styles/theme/theme';
import { Data, DataSchema } from '../../../../utils/data';
import ExecutionStatus, {
  getArtifactExecStateAsTableRow,
  isFailed,
  isInitial,
  isLoading,
} from '../../../../utils/shared';
import { StatusIndicator } from '../../workflowStatus';

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
      <Alert style={{ marginTop: '10px' }} severity="error">
        <AlertTitle>Failed to load historical data.</AlertTitle>
        <pre>{historyWithLoadingStatus.status.err}</pre>
      </Alert>
    );
  }

  const historicalData: Data = {
    schema: metricHistorySchema,
    data: (historyWithLoadingStatus.results?.results ?? []).map(
      (artifactStatusResult) => {
        return getArtifactExecStateAsTableRow(artifactStatusResult);
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

  const timestamps = dataToPlot.map((x) => new Date(x['timestamp'] as string));
  const values = dataToPlot.map((x) => x['value']);

  return (
    <Box display="flex" justifyContent="center" flexDirection="column">
      {dataToPlot.length > 0 && (
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
              margin: { b: 0, t: 0, l: 0, r: 0, pad: 8 },
              xaxis: { automargin: true, type: 'date' },
              yaxis: { automargin: true, ticksuffix: ' ' }
            }}
          />
        </Box>
      )}

      <Box mt="32px">
        {dataSortedByLatest.map((entry, index) => {
          let backgroundColor, hoverColor;
          if (entry.status === ExecutionStatus.Succeeded) {
            backgroundColor = theme.palette.green[100];
            hoverColor = theme.palette.green[200];
          } else if (entry.status === ExecutionStatus.Failed) {
            backgroundColor = theme.palette.red[25];
            hoverColor = theme.palette.red[100];
          } else {
            backgroundColor = theme.palette.gray[100];
            hoverColor = theme.palette.gray[200];
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
                width: 'auto',
              }}
            >
              <Box sx={{ flex: 1, display: 'flex', alignItems: 'center' }}>
                <StatusIndicator status={entry.status as ExecutionStatus} />

                <Typography sx={{ ml: 1 }} variant="body2">
                  {entry.timestamp}
                </Typography>
              </Box>
              <Typography variant="body1">
                {entry.value ? entry.value.toString() : '-'}
              </Typography>
            </Box>
          );
        })}
      </Box>
    </Box>
  );
};

export default MetricsHistory;
