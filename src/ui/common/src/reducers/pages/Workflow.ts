import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export type PerDagPageState = {
  // TODO: Add selected nodes, etc. here
};

type PerWorkflowPageState = {
  dagId: string;
  dagResultId?: string;
  perDagPageStates: {
    [dagOrResultId: string]: PerDagPageState;
  };
};

export type WorkflowPageState = {
  workflowId?: string;
  perWorkflowPageStates: {
    [workflowId: string]: PerWorkflowPageState;
  };
};

const initialState: WorkflowPageState = { perWorkflowPageStates: {} };
function initializePerWorkflowPageStateIfNotExists(
  state: WorkflowPageState,
  workflowId: string,
  dagId: string,
) {
  if (!state.perWorkflowPageStates[workflowId]) {
    state.perWorkflowPageStates[workflowId] = { dagId, perDagPageStates: {} };
  }
}

export const workflowPageSlice = createSlice({
  name: 'workflowPage',
  initialState,
  reducers: {
    selectDag: (state, { payload }: PayloadAction<{ workflowId: string; dagId: string }>) => {
      initializePerWorkflowPageStateIfNotExists(state, payload.workflowId, payload.dagId);
      state.perWorkflowPageStates[payload.workflowId].perDagPageStates[payload.dagId] = {}
    },
    selectDagResult: (
      state,
      { payload }: PayloadAction<{ workflowId: string; dagId: string; dagResultId: string }>
    ) => {
      initializePerWorkflowPageStateIfNotExists(state, payload.workflowId, payload.dagId);
      state.perWorkflowPageStates[payload.workflowId].dagResultId =
        payload.dagResultId;
      if (
        !state.perWorkflowPageStates[payload.workflowId].perDagPageStates[
        payload.dagResultId
        ]
      ) {
        state.perWorkflowPageStates[payload.workflowId].perDagPageStates[
          payload.dagResultId
        ] = {};
      }
    },
  },
});

export const { selectDag, selectDagResult } = workflowPageSlice.actions;
export default workflowPageSlice.reducer;
