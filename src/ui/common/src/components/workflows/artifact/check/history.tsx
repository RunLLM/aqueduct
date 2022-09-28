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

type CheckHistoryProps = {
    historyWithLoadingStatus?: ArtifactResultsWithLoadingStatus;
    checkLevel?: string;
}

// TODO: Bring over data schema from check details page
const checkHistorySchema: DataSchema = {
    fields: [
        { name: 'status', type: 'varchar' },
        { name: 'level', type: 'varchar' },
        { name: 'value', type: 'varchar' },
        { name: 'timestamp', type: 'varchar' }
    ],
    pandas_version: '0.0.1', // Not sure what actual value to put here, just filling in for now :)
};

const CheckHistory: React.FC<CheckHistoryProps> = ({ historyWithLoadingStatus, checkLevel }) => {
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
                console.log('artifactStatusResult', artifactStatusResult);
                return {
                    status: artifactStatusResult.exec_state?.status ?? 'Unknown',
                    level: checkLevel ? checkLevel : 'undefined',
                    value: artifactStatusResult.content_serialized,
                    timestamp: artifactStatusResult.exec_state?.timestamps?.finished_at,
                };
            }
        ),
    };

    return (
        <Box display="flex" justifyContent="center" flexDirection="column">
            <PaginatedTable data={historicalData} />
        </Box>
    )
};

export default CheckHistory;
