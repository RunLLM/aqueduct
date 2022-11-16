import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';

import { apiAddress } from '../components/hooks/useAqueductConsts';
import { RootState } from '../stores/store';
import { Integration } from '../utils/integrations';

export interface IntegrationState {
  thunkState: string;
  integrations: { [id: string]: Integration };
}

const initialState: IntegrationState = {
  integrations: {},
  thunkState: 'IDLE',
};

export const handleLoadIntegrations = createAsyncThunk<
  { [id: string]: Integration },
  { apiKey: string; forceLoad?: boolean },
  { state: RootState }
>(
  'integrations/load',
  async (
    args: {
      apiKey: string;
      forceLoad?: boolean;
    },
    thunkAPI
  ) => {
    // The integrations are already defined, so just ignore this call if not force load.
    const state = thunkAPI.getState();
    const integrations = state.integrationsReducer.integrations;
    if (
      integrations &&
      Object.values(integrations).length > 0 &&
      !args.forceLoad
    ) {
      return integrations;
    }

    const { apiKey } = args;
    const response = await fetch(`${apiAddress}/api/integrations`, {
      method: 'GET',
      headers: {
        'api-key': apiKey,
      },
    });

    const responseBody = await response.json();
    if (!response.ok) {
      return thunkAPI.rejectWithValue(responseBody.error);
    }

    const integrationList = responseBody as Integration[];
    const result: { [id: string]: Integration } = {};
    integrationList.forEach(
      (integration) => (result[integration.id] = integration)
    );
    return result;
  }
);

export const integrationsSlice = createSlice({
  name: 'integrationsReducer',
  initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder.addCase(handleLoadIntegrations.fulfilled, (state, { payload }) => {
      state.integrations = payload;
    });
  },
});

export default integrationsSlice.reducer;
