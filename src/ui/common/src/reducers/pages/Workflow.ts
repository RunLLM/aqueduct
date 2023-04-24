import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export type PerDagPageState = {
  // TODO: Add selected nodes, etc. here
};

type PerWorkflowPageState = {
  perDagPageStates: {
    [dagOrResultId: string]: PerDagPageState;
  };
};

export type WorkflowPageState = {
  perWorkflowPageStates: {
    [workflowId: string]: PerWorkflowPageState;
  };
};

const initialState: WorkflowPageState = { perWorkflowPageStates: {} };
function initializePerWorkflowPageStateIfNotExists(
  state: WorkflowPageState,
  workflowId: string
) {
  if (!state.perWorkflowPageStates[workflowId]) {
    state.perWorkflowPageStates[workflowId] = { perDagPageStates: {} };
  }
}

export const workflowPageSlice = createSlice({
  name: 'workflowPage',
  initialState,
  reducers: {
    initializeDagOrResultPageIfNotExists: (
      state,
      {
        payload,
      }: PayloadAction<{
        workflowId: string;
        dagId: string;
        dagResultId?: string;
      }>
    ) => {
      initializePerWorkflowPageStateIfNotExists(state, payload.workflowId);
      const pageKey = payload.dagResultId ?? payload.dagId;

      if (
        !state.perWorkflowPageStates[payload.workflowId].perDagPageStates[
          pageKey
        ]
      ) {
        state.perWorkflowPageStates[payload.workflowId].perDagPageStates[
          pageKey
        ] = {};
      }
    },
  },
});

export const { initializeDagOrResultPageIfNotExists } =
  workflowPageSlice.actions;
export default workflowPageSlice.reducer;
