import { createAsyncThunk } from '@reduxjs/toolkit';

import { apiAddress } from '../components/hooks/useAqueductConsts';
import {
  GetArtifactResultResponse,
  SerializationType,
} from '../utils/artifacts';

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
          'metadata-only': 'false',
        },
      }
    );

    if (!res.ok) {
      const body = await res.json();
      return thunkAPI.rejectWithValue(body.error);
    }

    const formData = await res.formData();
    const metadataJson = await (formData.get('metadata') as File).text();
    const artifactResult = JSON.parse(
      metadataJson
    ) as GetArtifactResultResponse;

    if (formData.has('data')) {
      if (
        artifactResult.serialization_type === SerializationType.String ||
        artifactResult.serialization_type === SerializationType.Table ||
        artifactResult.serialization_type === SerializationType.BsonTable ||
        artifactResult.serialization_type === SerializationType.Json
      ) {
        artifactResult.data = await (formData.get('data') as File).text();
      } else if (
        artifactResult.serialization_type === SerializationType.Image
      ) {
        // We first convert the image bytes into a base64 encoded string.
        const toBase64 = (file) =>
          new Promise<string>((resolve, reject) => {
            const reader = new FileReader();
            reader.readAsDataURL(file);
            reader.onload = () =>
              resolve(
                // Use a regex to remove the data url part.
                (reader.result as string)
                  .replace('data:', '')
                  .replace(/^.+,/, '')
              );
            reader.onerror = (error) => reject(error);
          });

        artifactResult.data = await toBase64(formData.get('data') as File);
      }
      return artifactResult.data;
    }

    return undefined;
  }
);
