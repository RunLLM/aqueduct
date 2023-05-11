import { createSlice, PayloadAction } from '@reduxjs/toolkit';

import { NodesResponse } from '../../handlers/responses/node';

// For now, this is an empty object.
export type PerDagPageState = Record<string, never>;

export type NodeSelection = {
  nodeType: keyof NodesResponse;
  nodeId: string;
};

type PerWorkflowPageState = {
  SelectedNode?: NodeSelection;
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
    selectNode: (
      state,
      {
        payload,
      }: PayloadAction<{ workflowId: string; selection?: NodeSelection }>
    ) => {
      initializePerWorkflowPageStateIfNotExists(state, payload.workflowId);
      state.perWorkflowPageStates[payload.workflowId].SelectedNode =
        payload.selection;
    },
  },
});

export const { initializeDagOrResultPageIfNotExists, selectNode } =
  workflowPageSlice.actions;
export default workflowPageSlice.reducer;
