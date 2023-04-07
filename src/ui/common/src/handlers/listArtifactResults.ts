import { createAsyncThunk } from '@reduxjs/toolkit';

import { apiAddress } from '../components/hooks/useAqueductConsts';
import { ListArtifactResultsResponse } from './responses/artifactDeprecated';

export const handleListArtifactResults = createAsyncThunk<
  ListArtifactResultsResponse,
  {
    apiKey: string;
    workflowId: string;
    artifactId: string;
  }
>(
  'api/list_artifact_results',
  async (
    args: {
      apiKey: string;
      workflowId: string;
      artifactId: string;
    },
    thunkAPI
  ) => {
    const { apiKey, workflowId, artifactId } = args;
    const resp = await fetch(
      `${apiAddress}/api/workflow/${workflowId}/artifact/${artifactId}/results`,
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

    return body as ListArtifactResultsResponse;
  }
);
