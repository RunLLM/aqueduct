// Before we fully migrate workflow result page, this store is
// only used for operator and artifact details page.
import { createSlice } from '@reduxjs/toolkit';

import { handleGetWorkflowDag } from '../handlers/getWorkflowDag';
import { DagResponse } from '../handlers/responses/dag';
import { LoadingStatus, LoadingStatusEnum } from '../utils/shared';

export type WorkflowDagWithLoadingStatus = {
  status: LoadingStatus;
  result?: DagResponse;
};

export interface WorkflowDagsState {
  results: {
    [id: string]: WorkflowDagWithLoadingStatus;
  };
}

const initialState: WorkflowDagsState = { results: {} };

export const workflowDagsSlice = createSlice({
  name: 'workflowDagsReducer',
  initialState: initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder.addCase(handleGetWorkflowDag.pending, (state, { meta }) => {
      const id = meta.arg.workflowDagId;
      state.results[id] = {
        status: { loading: LoadingStatusEnum.Loading, err: '' },
      };
    });
    builder.addCase(
      handleGetWorkflowDag.fulfilled,
      (state, { meta, payload }) => {
        const id = meta.arg.workflowDagId;
        state.results[id] = {
          result: payload,
          status: { loading: LoadingStatusEnum.Succeeded, err: '' },
        };
      }
    );
    builder.addCase(
      handleGetWorkflowDag.rejected,
      (state, { meta, payload }) => {
        const id = meta.arg.workflowDagId;

        state.results[id] = {
          status: { loading: LoadingStatusEnum.Failed, err: payload as string },
        };
      }
    );
  },
});

export default workflowDagsSlice.reducer;
