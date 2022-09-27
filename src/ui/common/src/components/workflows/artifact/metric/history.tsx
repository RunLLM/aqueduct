import { CircularProgress } from '@mui/material';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import React from 'react';
import Plot from 'react-plotly.js';

import PaginatedTable from '../../../../components/tables/PaginatedTable';
import { ArtifactResultsWithLoadingStatus } from '../../../../reducers/artifactResults';
import { theme } from '../../../../styles/theme/theme';
import { Data, DataSchema } from '../../../../utils/data';
import { isFailed, isInitial, isLoading } from '../../../../utils/shared';

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
          timestamp: artifactStatusResult.exec_state?.timestamps?.finished_at,
          value: artifactStatusResult.content_serialized,
        };
      }
    ),
  };

  const dataToPlot = historicalData.data.filter(
    (x) => !!x['timestamp'] && !!x['value']
  );
  const timestamps = dataToPlot.map((x) => x['timestamp']);
  const values = dataToPlot.map((x) => x['value']);
  return (
    <Box display="flex" justifyContent="center" flexDirection="column">
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
        layout={{ width: '100%', height: '100%' }}
      />
      <PaginatedTable data={historicalData} />
    </Box>
  );
};

export default MetricsHistory;
