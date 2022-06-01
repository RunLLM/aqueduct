import { useAqueductConsts } from '@aqueducthq/common/src/components/hooks/useAqueductConsts';
import { Integration } from '@aqueducthq/common/src/utils/integrations';
import { createAsyncThunk, createSlice, PayloadAction } from '@reduxjs/toolkit';
import { RootState } from '@stores/store';

export interface IntegrationState {
    thunkState: string;
    integrations: Integration[];
}

const initialState: IntegrationState = {
    integrations: [],
    thunkState: 'IDLE',
};

const { httpProtocol, apiAddress } = useAqueductConsts();
export const handleLoadIntegrations = createAsyncThunk<
    Integration[],
    { apiKey: string; forceLoad?: boolean },
    { state: RootState }
>(
    'integrations/load',
    async (
        args: {
            apiKey: string;
            forceLoad?: boolean;
        },
        thunkAPI,
    ) => {
        // The integrations are already defined, so just ignore this call if not force load.
        const state = thunkAPI.getState();
        const integrations = state.integrationsReducer.integrations;
        if (integrations && integrations.length > 0 && !args.forceLoad) {
            return integrations;
        }

        const { apiKey } = args;
        const response = await fetch(`${httpProtocol}://${apiAddress}/integrations`, {
            method: 'GET',
            headers: {
                'api-key': apiKey,
            },
        });

        const responseBody = await response.json();
        if (!response.ok) {
            return thunkAPI.rejectWithValue(responseBody.error);
        }

        return responseBody;
    },
);

const handleSetIntegrationsReducer = (state: IntegrationState, payload: Integration[]) => {
    state.integrations = payload;
};

export const integrationsSlice = createSlice({
    name: 'integrationsReducer',
    initialState,
    reducers: {},
    extraReducers: (builder) => {
        builder.addCase(handleLoadIntegrations.fulfilled, (state, { payload }: PayloadAction<Integration[]>) => {
            handleSetIntegrationsReducer(state, payload);
        });
    },
});

export default integrationsSlice.reducer;
