// Before we fully migrate workflow result page, this store is
// only used for metrics and check details page.
import { createSlice } from '@reduxjs/toolkit';

import { handleListArtifactResults } from '../handlers/listArtifactResults';
import { ListArtifactResultsResponse } from '../handlers/responses/artifact';
import { LoadingStatus, LoadingStatusEnum } from '../utils/shared';

export interface WorkflowDagResultsState {
  artifacts: {
    [id: string]: {
      status: LoadingStatus;
      results?: ListArtifactResultsResponse;
    };
  };
}

const initialState: WorkflowDagResultsState = { artifacts: {} };

export const artifactResultsSlice = createSlice({
  name: 'artifactResultsReducer',
  initialState: initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder.addCase(handleListArtifactResults.pending, (state, { meta }) => {
      const id = meta.arg.artifactId;
      state.artifacts[id] = {
        status: { loading: LoadingStatusEnum.Loading, err: '' },
      };
    });
    builder.addCase(
      handleListArtifactResults.fulfilled,
      (state, { meta, payload }) => {
        const id = meta.arg.artifactId;
        state.artifacts[id] = {
          results: payload,
          status: { loading: LoadingStatusEnum.Succeeded, err: '' },
        };
      }
    );
    builder.addCase(
      handleListArtifactResults.rejected,
      (state, { meta, payload }) => {
        const id = meta.arg.artifactId;

        state.artifacts[id] = {
          status: { loading: LoadingStatusEnum.Failed, err: payload as string },
        };
      }
    );
  },
});

export default artifactResultsSlice.reducer;
