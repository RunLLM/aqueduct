// Before we fully migrate workflow result page, this store is
// only used for operator and artifact details page.
import { createSlice } from '@reduxjs/toolkit';

import { handleGetWorkflowDagResult } from '../handlers/getWorkflowDagResult';
import { DagResultResponse } from '../handlers/responses/dag';
import { LoadingStatus, LoadingStatusEnum } from '../utils/shared';

export interface WorkflowDagResultsState {
  results: {
    [id: string]: {
      status: LoadingStatus;
      result?: DagResultResponse;
    };
  };
}

const initialState: WorkflowDagResultsState = { results: {} };

export const workflowDagResultsSlice = createSlice({
  name: 'workflowDagResultsReducer',
  initialState: initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder.addCase(handleGetWorkflowDagResult.pending, (state, { meta }) => {
      const id = meta.arg.workflowDagResultId;
      state.results[id] = {
        status: { loading: LoadingStatusEnum.Loading, err: '' },
      };
    });
    builder.addCase(
      handleGetWorkflowDagResult.fulfilled,
      (state, { meta, payload }) => {
        const id = meta.arg.workflowDagResultId;
        state.results[id] = {
          result: payload,
          status: { loading: LoadingStatusEnum.Succeeded, err: '' },
        };
      }
    );
    builder.addCase(
      handleGetWorkflowDagResult.rejected,
      (state, { meta, payload }) => {
        const id = meta.arg.workflowDagResultId;

        state.results[id] = {
          status: { loading: LoadingStatusEnum.Failed, err: payload as string },
        };
      }
    );
  },
});

export default workflowDagResultsSlice.reducer;
