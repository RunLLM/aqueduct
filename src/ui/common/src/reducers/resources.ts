import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';

import { apiAddress } from '../components/hooks/useAqueductConsts';
import { RootState } from '../stores/store';
import { Resource } from '../utils/resources';
import { LoadingStatus, LoadingStatusEnum } from '../utils/shared';

export interface ResourceState {
  status: LoadingStatus;
  resources: { [id: string]: Resource };
}

const initialState: ResourceState = {
  resources: {},
  status: { loading: LoadingStatusEnum.Initial, err: '' },
};

export const handleLoadResources = createAsyncThunk<
  { [id: string]: Resource },
  { apiKey: string; forceLoad?: boolean },
  { state: RootState }
>(
  'resources/load',
  async (
    args: {
      apiKey: string;
      forceLoad?: boolean;
    },
    thunkAPI
  ) => {
    // The resources are already defined, so just ignore this call if not force load.
    const state = thunkAPI.getState();
    const resources = state.resourcesReducer.resources;
    if (
      resources &&
      Object.values(resources).length > 0 &&
      !args.forceLoad
    ) {
      return resources;
    }

    const { apiKey } = args;
    const response = await fetch(`${apiAddress}/api/resources`, {
      method: 'GET',
      headers: {
        'api-key': apiKey,
      },
    });

    const responseBody = await response.json();
    if (!response.ok) {
      return thunkAPI.rejectWithValue(responseBody.error);
    }

    const resourceList = responseBody as Resource[];
    const result: { [id: string]: Resource } = {};
    resourceList.forEach(
      (resource) => (result[resource.id] = resource)
    );
    return result;
  }
);

export const resourcesSlice = createSlice({
  name: 'resourcesReducer',
  initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder.addCase(handleLoadResources.pending, (state) => {
      state.status = { loading: LoadingStatusEnum.Loading, err: '' };
    });
    builder.addCase(handleLoadResources.fulfilled, (state, { payload }) => {
      state.resources = payload;
      state.status = { loading: LoadingStatusEnum.Succeeded, err: '' };
    });
    builder.addCase(handleLoadResources.rejected, (state, { payload }) => {
      state.status = {
        loading: LoadingStatusEnum.Failed,
        err: payload as string,
      };
    });
  },
});

export default resourcesSlice.reducer;
