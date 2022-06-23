import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';

import { useAqueductConsts } from '../components/hooks/useAqueductConsts';
import { RootState } from '../stores/store';
import ExecutionStatus from '../utils/shared';

type TableData = {
  status: ExecutionStatus;
  data: string;
  err: string;
};

type PreviewResponse = {
  data: string;
};

const initialState: Record<string, TableData> = {
  table: {
    status: ExecutionStatus.Pending,
    data: null,
    err: '',
  },
};

export const tableKeyFn = (table: string): string => `table${table}`;

const { apiAddress } = useAqueductConsts();
export const handleLoadIntegrationTable = createAsyncThunk<
  string,
  { apiKey: string; integrationId: string; table: string; forceLoad?: boolean },
  { state: RootState }
>(
  'integrationTableDataReducer/load',
  async (
    args: {
      apiKey: string;
      integrationId: string;
      table: string;
      forceLoad?: boolean;
    },
    thunkAPI
  ) => {
    const { apiKey, integrationId, table } = args;

    // The integrations are already defined, so just ignore this call if not force load.
    const state = thunkAPI.getState();
    const tableKey = tableKeyFn(table);
    if (
      table === '' ||
      (state.integrationTableDataReducer.hasOwnProperty(tableKey) &&
        state.integrationTableDataReducer[tableKey].data &&
        !args.forceLoad)
    ) {
      return state.integrationTableDataReducer[tableKey].data;
    }
    const tableResponse = await fetch(
      `${apiAddress}/api/integration/${integrationId}/preview_table`,
      {
        method: 'GET',
        headers: {
          'api-key': apiKey,
          'table-name': table,
        },
      }
    );

    const tableResponseBody = await tableResponse.json();

    if (!tableResponse.ok || !(tableResponseBody && tableResponseBody.data)) {
      return thunkAPI.rejectWithValue(tableResponseBody.error);
    } else {
      return (tableResponseBody as PreviewResponse).data;
    }
  }
);

export const integrationTableDataSlice = createSlice({
  name: 'integrationTableDataReducer',
  initialState: initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder.addCase(
      handleLoadIntegrationTable.rejected,
      (state, { meta, payload }) => {
        const table = meta.arg.table;
        const tableKey = tableKeyFn(table);

        state[tableKey] = {
          data: null,
          err: payload as string,
          status: ExecutionStatus.Failed,
        };
      }
    );
    builder.addCase(
      handleLoadIntegrationTable.fulfilled,
      (state, { meta, payload }) => {
        const table = meta.arg.table;
        const tableKey = tableKeyFn(table);

        state[tableKey] = {
          data: payload,
          err: '',
          status: ExecutionStatus.Succeeded,
        };
      }
    );
  },
});

export default integrationTableDataSlice.reducer;
