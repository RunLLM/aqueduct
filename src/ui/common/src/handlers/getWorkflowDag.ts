import { createAsyncThunk } from '@reduxjs/toolkit';

import { apiAddress } from '../components/hooks/useAqueductConsts';
import { DagResponse } from './responses/dagDeprecated';

export const handleGetWorkflowDag = createAsyncThunk<
  DagResponse,
  {
    apiKey: string;
    workflowId: string;
    workflowDagId: string;
  }
>(
  'api/get_workflow_dag',
  async (
    args: {
      apiKey: string;
      workflowId: string;
      workflowDagId: string;
    },
    thunkAPI
  ) => {
    const { apiKey, workflowId, workflowDagId } = args;
    const resp = await fetch(
      `${apiAddress}/api/workflow/${workflowId}/dag/${workflowDagId}`,
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

    return body as DagResponse;
  }
);
