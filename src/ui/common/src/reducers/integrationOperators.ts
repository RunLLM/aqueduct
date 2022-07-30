import { createAsyncThunk, createSlice, PayloadAction } from '@reduxjs/toolkit';

import { useAqueductConsts } from '../components/hooks/useAqueductConsts';
import { RootState } from '../stores/store';
import { Operator } from '../utils/operators';
import { LoadingStatus, LoadingStatusEnum } from '../utils/shared';

export type OperatorsForIntegrationItem = {
  operator: Operator;
  workflow_id: string;
  workflow_dag_id: string;
  is_active: boolean;
};

type OperatorsForIntegrationResponse = {
  operator_with_ids: OperatorsForIntegrationItem[];
};

export interface IntegrationOperatorsState {
  loadingStatus: LoadingStatus;
  operators: OperatorsForIntegrationItem[];
}

const initialState: IntegrationOperatorsState = {
  loadingStatus: { loading: LoadingStatusEnum.Initial, err: '' },
  operators: [],
};

const { apiAddress } = useAqueductConsts();
export const handleLoadIntegrationOperators = createAsyncThunk<
  OperatorsForIntegrationItem[],
  { apiKey: string; integrationId: string },
  { state: RootState }
>(
  'integrationOperators/load',
  async (
    args: {
      apiKey: string;
      integrationId: string;
    },
    thunkAPI
  ) => {
    const { apiKey, integrationId } = args;
    const response = await fetch(
      `${apiAddress}/api/integration/${integrationId}/operators`,
      {
        method: 'GET',
        headers: {
          'api-key': apiKey,
        },
      }
    );

    const responseBody = await response.json();
    if (!response.ok) {
      return thunkAPI.rejectWithValue(responseBody.error);
    }
    return (responseBody as OperatorsForIntegrationResponse).operator_with_ids;
  }
);

export const integrationOperatorsSlice = createSlice({
  name: 'integrationTablesReducer',
  initialState: initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder.addCase(handleLoadIntegrationOperators.pending, (state) => {
      state.loadingStatus = { loading: LoadingStatusEnum.Loading, err: '' };
    });
    builder.addCase(
      handleLoadIntegrationOperators.fulfilled,
      (state, { payload }: PayloadAction<OperatorsForIntegrationItem[]>) => {
        state.loadingStatus = { loading: LoadingStatusEnum.Succeeded, err: '' };
        state.operators = payload;
      }
    );
    builder.addCase(
      handleLoadIntegrationOperators.rejected,
      (state, { payload }) => {
        state.loadingStatus = {
          loading: LoadingStatusEnum.Failed,
          err: payload as string,
        };
      }
    );
  },
});

export default integrationOperatorsSlice.reducer;
