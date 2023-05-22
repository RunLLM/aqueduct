import { createSlice, PayloadAction } from '@reduxjs/toolkit';

import { NodesResponse } from '../../handlers/responses/node';

export type NodeSelection = {
  nodeType: keyof NodesResponse;
  nodeId: string;
};

type PerWorkflowPageState = {
  SelectedNode?: NodeSelection;
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
    state.perWorkflowPageStates[workflowId] = {};
  }
}

export const workflowPageSlice = createSlice({
  name: 'workflowPage',
  initialState,
  reducers: {
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

export const { selectNode } = workflowPageSlice.actions;
export default workflowPageSlice.reducer;
