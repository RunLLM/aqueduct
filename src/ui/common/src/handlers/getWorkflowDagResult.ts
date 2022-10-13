import { createAsyncThunk } from '@reduxjs/toolkit';

import { useAqueductConsts } from '../components/hooks/useAqueductConsts';
import { DagResultResponse } from './responses/dag';

const { apiAddress } = useAqueductConsts();

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
    console.log('handleGetWorklfowDagResult body: ', body);

    if (!resp.ok) {
      console.log('error handleGetWorkflowDagResult: ', body.error);
      return thunkAPI.rejectWithValue(body.error);
    }

    return body as DagResultResponse;
  }
);
