import { createAsyncThunk } from '@reduxjs/toolkit';

import { DagResultResponse } from './responses/dag';

export const handleGetWorkflowDagResult = createAsyncThunk<
  DagResultResponse,
  {
    apiAddress: string;
    apiKey: string;
    workflowId: string;
    workflowDagResultId: string;
  }
>(
  'api/get_workflow_dag_result',
  async (
    args: {
      apiAddress: string;
      apiKey: string;
      workflowId: string;
      workflowDagResultId: string;
    },
    thunkAPI
  ) => {
    const { apiAddress, apiKey, workflowId, workflowDagResultId } = args;
    const resp = await fetch(
      `${apiAddress}/api/workflow/${workflowId}/result/${workflowDagResultId}`,
      {
        method: 'GET',
        headers: {
          'api-key': apiKey,
        },
      }
    );

    const body = await resp.json();

    if (!resp.ok) {
      return thunkAPI.rejectWithValue(body.error);
    }

    return body as DagResultResponse;
  }
);
