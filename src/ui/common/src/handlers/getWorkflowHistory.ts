import { createAsyncThunk } from '@reduxjs/toolkit';

import { apiAddress } from '../components/hooks/useAqueductConsts';
import { WorkflowHistoryResponse } from './responses/workflowHistoryDeprecated';

export const handleGetWorkflowHistory = createAsyncThunk<
  WorkflowHistoryResponse,
  {
    apiKey: string;
    workflowId: string;
  }
>(
  'api/get_workflow_history',
  async (
    args: {
      apiKey: string;
      workflowId: string;
    },
    thunkAPI
  ) => {
    const { apiKey, workflowId } = args;
    const resp = await fetch(
      `${apiAddress}/api/workflow/${workflowId}/history`,
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

    return body as WorkflowHistoryResponse;
  }
);
