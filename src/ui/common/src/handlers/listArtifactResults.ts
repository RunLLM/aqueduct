import { createAsyncThunk } from '@reduxjs/toolkit';

import { ListArtifactResultsResponse } from './responses/artifact';

export const handleListArtifactResults = createAsyncThunk<
  ListArtifactResultsResponse,
  {
    apiAddress: string;
    apiKey: string;
    artifactId: string;
  }
>(
  'api/list_artifact_results',
  async (
    args: {
      apiAddress: string;
      apiKey: string;
      artifactId: string;
    },
    thunkAPI
  ) => {
    const { apiAddress, apiKey, artifactId } = args;
    const resp = await fetch(
      `${apiAddress}/api/artifact/${artifactId}/results`,
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
