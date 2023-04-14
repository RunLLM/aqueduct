import { createAsyncThunk } from '@reduxjs/toolkit';

import { apiAddress } from '../components/hooks/useAqueductConsts';
import { DagResultResponse } from './responses/dagDeprecated';

export const handleGetWorkflowDagResult = createAsyncThunk<
  DagResultResponse,
  {
    apiKey: string;
    workflowId: string;
    workflowDagResultId: string;
  }
>(
  'api/get_workflow_dag_result',
  async (
    args: {
      apiKey: string;
      workflowId: string;
      workflowDagResultId: string;
    },
    thunkAPI
  ) => {
    const { apiKey, workflowId, workflowDagResultId } = args;
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
