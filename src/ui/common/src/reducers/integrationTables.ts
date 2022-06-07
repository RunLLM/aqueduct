import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';

import { useAqueductConsts } from '../components/hooks/useAqueductConsts';
import { RootState } from '../stores/store';
import ExecutionStatus from '@utils/shared';

export interface IntegrationTablesState {
    integrationTables: string[];
    thunkState: ExecutionStatus;
    err: string;
}

type DiscoverResponse = {
    table_names: string[];
};

const initialState: IntegrationTablesState = {
    integrationTables: [],
    thunkState: ExecutionStatus.Pending,
    err: '',
};

const { httpProtocol, apiAddress } = useAqueductConsts();
export const handleLoadIntegrationTables = createAsyncThunk<
    any | string[],
    { apiKey: string; integrationId: string; forceLoad?: boolean },
    { state: RootState }
>(
    'integrationTablesReducer/load',
    async (
        args: {
            apiKey: string;
            integrationId: string;
            forceLoad?: boolean;
        },
        thunkAPI,
    ) => {
        // The integrations are already defined, so just ignore this call if not force load.
        const state = thunkAPI.getState();
        const integrationTables = state.integrationTablesReducer.integrationTables;
        if (integrationTables && integrationTables.length > 0 && !args.forceLoad) {
            return integrationTables;
        }

        const { apiKey, integrationId } = args;
        const response = await fetch(`${httpProtocol}://${apiAddress}/integration/${integrationId}/discover`, {
            method: 'GET',
            headers: {
                'api-key': apiKey,
            },
        });

        const responseBody = await response.json();
        if (!response.ok) {
            return thunkAPI.rejectWithValue(responseBody.error);
        }
        return (responseBody as DiscoverResponse).table_names;
    },
);

export const integrationTablesSlice = createSlice({
    name: 'integrationTablesReducer',
    initialState,
    reducers: {},
    extraReducers: (builder) => {
        builder.addCase(handleLoadIntegrationTables.pending, (state) => {
            state.thunkState = ExecutionStatus.Pending;
        });
        builder.addCase(handleLoadIntegrationTables.rejected, (state, { payload }) => {
            state.thunkState = ExecutionStatus.Failed;
            state.err = payload as string;
        });
        builder.addCase(handleLoadIntegrationTables.fulfilled, (state, { payload }) => {
            state.thunkState = ExecutionStatus.Succeeded;
            state.integrationTables = payload;
        });
    },
});

export default integrationTablesSlice.reducer;