import { createAsyncThunk, createSlice, PayloadAction } from '@reduxjs/toolkit';

import { useAqueductConsts } from '../components/hooks/useAqueductConsts';
import { RootState } from '../stores/store';
import { Data } from '../utils/data';
import { Operator } from '../utils/operators';
import { LoadingStatus, LoadingStatusEnum } from '../utils/shared';

export type OperatorsForIntegrationItem = {
  operator: Operator;
  workflow_id: string;
  workflow_dag_id: string;
  is_active: boolean;
};

type OperatorsForIntegrationResponse = {
  operator_with_ids: OperatorsForIntegrationItem[];
};

type IntegrationOperatorsState = {
  status: LoadingStatus;
  operators: OperatorsForIntegrationItem[];
};

type ObjectPreviewResponse = {
  data: string;
};

export type ObjectState = {
  status: LoadingStatus;
  data?: Data;
};

type ListObjectsState = {
  names: string[];
  status: LoadingStatus;
};

type DiscoverResponse = {
  table_names: string[];
};

export interface IntegrationState {
  connectionStatus: LoadingStatus;
  deletionStatus: LoadingStatus;
  operators: IntegrationOperatorsState;
  objectNames: ListObjectsState;
  objects: Record<string, ObjectState>;
}

const initialState: IntegrationState = {
  connectionStatus: { loading: LoadingStatusEnum.Initial, err: '' },
  deletionStatus: { loading: LoadingStatusEnum.Initial, err: '' },
  operators: {
    status: { loading: LoadingStatusEnum.Initial, err: '' },
    operators: [],
  },
  objectNames: {
    status: { loading: LoadingStatusEnum.Initial, err: '' },
    names: [],
  },
  objects: {},
};

const { apiAddress } = useAqueductConsts();
export const handleLoadIntegrationOperators = createAsyncThunk<
  OperatorsForIntegrationItem[],
  { apiKey: string; integrationId: string },
  { state: RootState }
>(
  'integration/loadOperators',
  async (
    args: {
      apiKey: string;
      integrationId: string;
    },
    thunkAPI
  ) => {
    const { apiKey, integrationId } = args;
    const response = await fetch(
      `${apiAddress}/api/integration/${integrationId}/operators`,
      {
        method: 'GET',
        headers: {
          'api-key': apiKey,
        },
      }
    );

    const responseBody = await response.json();
    if (!response.ok) {
      return thunkAPI.rejectWithValue(responseBody.error);
    }
    return (responseBody as OperatorsForIntegrationResponse).operator_with_ids;
  }
);

export const objectKeyFn = (object: string): string => `object${object}`;
export const handleLoadIntegrationObject = createAsyncThunk<
  Data,
  {
    apiKey: string;
    integrationId: string;
    object: string;
    forceLoad?: boolean;
  },
  { state: RootState }
>(
  'integration/loadObject',
  async (
    args: {
      apiKey: string;
      integrationId: string;
      object: string;
      forceLoad?: boolean;
    },
    thunkAPI
  ) => {
    const { apiKey, integrationId, object } = args;

    // The integrations are already defined, so just ignore this call if not force load.
    const state = thunkAPI.getState();
    const objectKey = objectKeyFn(object);
    if (
      !args.forceLoad &&
      (object === '' ||
        (state.integrationReducer.objects.hasOwnProperty(objectKey) &&
          state.integrationReducer.objects[objectKey].data))
    ) {
      return state.integrationReducer.objects[objectKey].data;
    }

    const objectResponse = await fetch(
      `${apiAddress}/api/integration/${integrationId}/preview_table`,
      {
        method: 'GET',
        headers: {
          'api-key': apiKey,
          'table-name': object,
        },
      }
    );

    const objectResponseBody = await objectResponse.json();

    if (
      !objectResponse.ok ||
      !(objectResponseBody && objectResponseBody.data)
    ) {
      return thunkAPI.rejectWithValue(objectResponseBody.error);
    } else {
      const serialized = (objectResponseBody as ObjectPreviewResponse).data;
      return JSON.parse(serialized) as Data;
    }
  }
);

export const handleListIntegrationObjects = createAsyncThunk<
  any | string[],
  { apiKey: string; integrationId: string; forceLoad?: boolean },
  { state: RootState }
>(
  'integration/loadObjects',
  async (
    args: {
      apiKey: string;
      integrationId: string;
      forceLoad?: boolean;
    },
    thunkAPI
  ) => {
    // The integrations are already defined, so just ignore this call if not force load.
    const state = thunkAPI.getState();
    const objects = state.integrationReducer.objectNames.names;
    if (objects && objects.length > 0 && !args.forceLoad) {
      return objects;
    }

    const { apiKey, integrationId } = args;
    const response = await fetch(
      `${apiAddress}/api/integration/${integrationId}/discover`,
      {
        method: 'GET',
        headers: {
          'api-key': apiKey,
        },
      }
    );

    const responseBody = await response.json();
    if (!response.ok) {
      return thunkAPI.rejectWithValue(responseBody.error);
    }
    return (responseBody as DiscoverResponse).table_names;
  }
);

export const handleDeleteIntegration = createAsyncThunk<
  void,
  { apiKey: string; integrationId: string }
>(
  'integration/delete',
  async (
    args: {
      apiKey: string;
      integrationId: string;
      forceLoad?: boolean;
    },
    thunkAPI
  ) => {
    const { apiKey, integrationId } = args;
    const response = await fetch(
      `${apiAddress}/api/integration/${integrationId}/delete`,
      {
        method: 'POST',
        headers: {
          'api-key': apiKey,
        },
      }
    );

    const responseBody = await response.json();

    if (!response.ok) {
      return thunkAPI.rejectWithValue(responseBody.error);
    }
    // const { apiKey, integrationId } = args;
    // const response = await fetch(
    //   `${apiAddress}/api/integration/${integrationId}/delete`,
    //   {
    //     method: 'POST',
    //     headers: {
    //       'api-key': apiKey,
    //     },
    //   }
    // );

    // const responseBody = await response.json();

    // if (!response.ok) {
    //   return thunkAPI.rejectWithValue(responseBody.error);
    // }
  }
);

export const handleTestConnectIntegration = createAsyncThunk<
  void,
  { apiKey: string; integrationId: string }
>(
  'integration/testConnect',
  async (
    args: {
      apiKey: string;
      integrationId: string;
      forceLoad?: boolean;
    },
    thunkAPI
  ) => {
    const { apiKey, integrationId } = args;
    const response = await fetch(
      `${apiAddress}/api/integration/${integrationId}/test`,
      {
        method: 'POST',
        headers: {
          'api-key': apiKey,
        },
      }
    );

    const responseBody = await response.json();

    if (!response.ok) {
      return thunkAPI.rejectWithValue(responseBody.error);
    }
  }
);

export const integrationSlice = createSlice({
  name: 'integrationTablesReducer',
  initialState: initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder.addCase(handleLoadIntegrationObject.pending, (state, { meta }) => {
      const object = meta.arg.object;
      const objectKey = objectKeyFn(object);
      state.objects[objectKey] = {
        status: { loading: LoadingStatusEnum.Loading, err: '' },
      };
    });
    builder.addCase(
      handleLoadIntegrationObject.rejected,
      (state, { meta, payload }) => {
        const object = meta.arg.object;
        const objectKey = objectKeyFn(object);

        state.objects[objectKey] = {
          status: { loading: LoadingStatusEnum.Failed, err: payload as string },
        };
      }
    );
    builder.addCase(
      handleLoadIntegrationObject.fulfilled,
      (state, { meta, payload }) => {
        const object = meta.arg.object;
        const objectKey = objectKeyFn(object);

        state.objects[objectKey] = {
          data: payload,
          status: { loading: LoadingStatusEnum.Succeeded, err: '' },
        };
      }
    );
    builder.addCase(handleLoadIntegrationOperators.pending, (state) => {
      state.operators.status = { loading: LoadingStatusEnum.Loading, err: '' };
    });
    builder.addCase(
      handleLoadIntegrationOperators.fulfilled,
      (state, { payload }: PayloadAction<OperatorsForIntegrationItem[]>) => {
        state.operators.status = {
          loading: LoadingStatusEnum.Succeeded,
          err: '',
        };
        state.operators.operators = payload;
      }
    );
    builder.addCase(
      handleLoadIntegrationOperators.rejected,
      (state, { payload }) => {
        state.operators.status = {
          loading: LoadingStatusEnum.Failed,
          err: payload as string,
        };
      }
    );
    builder.addCase(handleListIntegrationObjects.pending, (state) => {
      state.objectNames.status = {
        loading: LoadingStatusEnum.Loading,
        err: '',
      };
    });
    builder.addCase(
      handleListIntegrationObjects.rejected,
      (state, { payload }) => {
        state.objectNames.status = {
          loading: LoadingStatusEnum.Failed,
          err: payload as string,
        };
      }
    );
    builder.addCase(
      handleListIntegrationObjects.fulfilled,
      (state, { payload }) => {
        state.objectNames.status = {
          loading: LoadingStatusEnum.Succeeded,
          err: '',
        };
        state.objectNames.names = payload;
      }
    );

    builder.addCase(handleDeleteIntegration.pending, (state) => {
      state.deletionStatus = { loading: LoadingStatusEnum.Loading, err: '' };
    });
    builder.addCase(handleDeleteIntegration.rejected, (state, { payload }) => {
      state.deletionStatus = {
        loading: LoadingStatusEnum.Failed,
        err: payload as string,
      };
    });
    builder.addCase(handleDeleteIntegration.fulfilled, (state) => {
      state.deletionStatus = {
        loading: LoadingStatusEnum.Succeeded,
        err: '',
      };
    });
    builder.addCase(handleTestConnectIntegration.pending, (state) => {
      state.connectionStatus = { loading: LoadingStatusEnum.Loading, err: '' };
    });
    builder.addCase(handleTestConnectIntegration.fulfilled, (state) => {
      state.connectionStatus = {
        loading: LoadingStatusEnum.Succeeded,
        err: '',
      };
    });
    builder.addCase(
      handleTestConnectIntegration.rejected,
      (state, { payload }) => {
        state.connectionStatus = {
          loading: LoadingStatusEnum.Failed,
          err: payload as string,
        };
      }
    );
  },
});

export default integrationSlice.reducer;
