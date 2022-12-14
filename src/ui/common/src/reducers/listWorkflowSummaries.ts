import { createAsyncThunk, createSlice, PayloadAction } from '@reduxjs/toolkit';

import { useAqueductConsts } from '../components/hooks/useAqueductConsts';
import { LoadingStatus, LoadingStatusEnum } from '../utils/shared';
import { ListWorkflowSummary } from '../utils/workflows';

export interface FetchWorkflowSummariesState {
  loadingStatus: LoadingStatus;
  workflows: ListWorkflowSummary[];
}

const initialWorkflowState: FetchWorkflowSummariesState = {
  loadingStatus: { loading: LoadingStatusEnum.Initial, err: '' },
  workflows: [],
};

export const handleFetchAllWorkflowSummaries = createAsyncThunk<
  ListWorkflowSummary[],
  { apiKey: string }
>(
  'listWorkflowReducer/fetch',
  async (
    args: {
      apiKey: string;
    },
    thunkAPI
  ) => {
    const { apiAddress } = useAqueductConsts();

    const { apiKey } = args;
    const response = await fetch(`${apiAddress}/api/workflows`, {
      method: 'GET',
      headers: {
        'api-key': apiKey,
      },
    });

    const body = await response.json();
    if (!response.ok) {
      return thunkAPI.rejectWithValue(body.error);
    }

    return body as ListWorkflowSummary[];
  }
);

export const listWorkflowSlice = createSlice({
  name: 'listWorkflowReducer',
  initialState: initialWorkflowState,
  reducers: {},
  extraReducers: (builder) => {
    builder.addCase(handleFetchAllWorkflowSummaries.pending, (state) => {
      state.loadingStatus = { loading: LoadingStatusEnum.Loading, err: '' };
    });

    builder.addCase(
      handleFetchAllWorkflowSummaries.fulfilled,
      (state, { payload }: PayloadAction<ListWorkflowSummary[]>) => {
        state.loadingStatus = { loading: LoadingStatusEnum.Succeeded, err: '' };
        state.workflows = payload;
      }
    );

    builder.addCase(
      handleFetchAllWorkflowSummaries.rejected,
      (state, { payload }) => {
        state.loadingStatus = {
          loading: LoadingStatusEnum.Failed,
          err: payload as string,
        };
      }
    );
  },
});

export default listWorkflowSlice.reducer;
