import { createSlice } from '@reduxjs/toolkit';

import { handleGetWorkflowHistory } from '../handlers/getWorkflowHistory';
import { WorkflowHistoryResponse } from '../handlers/responses/workflowHistory';
import { LoadingStatus, LoadingStatusEnum } from '../utils/shared';

// TODO(ENG-2680): Do we ever need to capture multiple workflows' histories at once?
export interface WorkflowHistoryState {
  status: LoadingStatus;
  history?: WorkflowHistoryResponse;
}

const initialState: WorkflowHistoryState = {
  status: { loading: LoadingStatusEnum.Initial, err: null },
  history: undefined,
};

export const workflowHistorySlice = createSlice({
  name: 'workflowHistoryReducer',
  initialState: initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder.addCase(handleGetWorkflowHistory.pending, (state, { meta }) => {
      state.status = {
        loading: LoadingStatusEnum.Loading,
        err: null,
      };
    }),
      builder.addCase(
        handleGetWorkflowHistory.fulfilled,
        (state, { meta, payload }) => {
          state.status = {
            loading: LoadingStatusEnum.Succeeded,
            err: null,
          };

          const workflowHistory = payload as WorkflowHistoryResponse;
          // Reverse this because the API returns oldest first.
          workflowHistory.versions.reverse();
          state.history = workflowHistory;
        }
      ),
      builder.addCase(
        handleGetWorkflowHistory.rejected,
        (state, { meta, payload }) => {
          state.status = {
            loading: LoadingStatusEnum.Failed,
            err: payload as string,
          };
        }
      );
  },
});

export default workflowHistorySlice.reducer;
