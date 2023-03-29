// Before we fully migrate workflow result page, this store is
// only used for metrics and check details page.
import { createSlice } from '@reduxjs/toolkit';

import { handleListArtifactResults } from '../handlers/listArtifactResults';
import { ListArtifactResultsResponse } from '../handlers/responses/artifactDeprecated';
import { LoadingStatus, LoadingStatusEnum } from '../utils/shared';

export type ArtifactResultsWithLoadingStatus = {
  status: LoadingStatus;
  results?: ListArtifactResultsResponse;
};

export interface ArtifactResultsState {
  artifacts: {
    [id: string]: ArtifactResultsWithLoadingStatus;
  };
}

const initialState: ArtifactResultsState = { artifacts: {} };

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
