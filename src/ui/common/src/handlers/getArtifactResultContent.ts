import { createAsyncThunk } from '@reduxjs/toolkit';

import { useAqueductConsts } from '../components/hooks/useAqueductConsts';

const { apiAddress } = useAqueductConsts();

export const handleGetArtifactResultContent = createAsyncThunk<
  string,
  {
    apiKey: string;
    artifactId: string;
    artifactResultId: string;
    workflowDagResultId: string;
  }
>(
  'api/get_artifact_result_content',
  async (
    args: {
      apiKey: string;
      artifactId: string;
      artifactResultId: string;
      workflowDagResultId: string;
    },
    thunkAPI
  ) => {
    const { apiKey, workflowDagResultId, artifactId } = args;
    const res = await fetch(
      `${apiAddress}/api/artifact/${workflowDagResultId}/${artifactId}/result`,
      {
        method: 'GET',
        headers: {
          'api-key': apiKey,
        },
      }
    );

    if (!res.ok) {
      const body = await res.json();
      return thunkAPI.rejectWithValue(body.error);
    }

    const formData = await res.formData();
    const data = await (formData.get('data') as File).text();

    return data;
  }
);
