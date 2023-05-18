import { createAsyncThunk, createSlice, PayloadAction } from '@reduxjs/toolkit';

import { apiAddress } from '../components/hooks/useAqueductConsts';
import { RootState } from '../stores/store';
import { Data, inferSchema, TableRow } from '../utils/data';
import { IntegrationConfig, Service } from '../utils/resources';
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
  connectNewStatus: LoadingStatus;
  editStatus: LoadingStatus;
  testConnectStatus: LoadingStatus;
  deletionStatus: LoadingStatus;
  operators: IntegrationOperatorsState;
  objectNames: ListObjectsState;
  objects: Record<string, ObjectState>;
}

const initialState: IntegrationState = {
  connectNewStatus: { loading: LoadingStatusEnum.Initial, err: '' },
  editStatus: { loading: LoadingStatusEnum.Initial, err: '' },
  testConnectStatus: { loading: LoadingStatusEnum.Initial, err: '' },
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

export const handleLoadIntegrationOperators = createAsyncThunk<
  OperatorsForIntegrationItem[],
  { apiKey: string; resourceId: string },
  { state: RootState }
>(
  'resource/loadOperators',
  async (
    args: {
      apiKey: string;
      resourceId: string;
    },
    thunkAPI
  ) => {
    const { apiKey, resourceId } = args;
    const response = await fetch(
      `${apiAddress}/api/resource/${resourceId}/operators`,
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
    resourceId: string;
    object: string;
    forceLoad?: boolean;
  },
  { state: RootState }
>(
  'resource/loadObject',
  async (
    args: {
      apiKey: string;
      resourceId: string;
      object: string;
      forceLoad?: boolean;
    },
    thunkAPI
  ) => {
    const { apiKey, resourceId, object } = args;

    // The resources are already defined, so just ignore this call if not force load.
    const state = thunkAPI.getState();
    const objectKey = objectKeyFn(object);
    if (
      !args.forceLoad &&
      (object === '' ||
        (state.resourceReducer.objects.hasOwnProperty(objectKey) &&
          state.resourceReducer.objects[objectKey].data))
    ) {
      return state.resourceReducer.objects[objectKey].data;
    }

    const objectResponse = await fetch(
      `${apiAddress}/api/resource/${resourceId}/preview`,
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
      const rawData = JSON.parse(serialized);

      // Distinguish between serialization types `table` vs `bson_table`,
      // since this information is not returned by backend.
      // TODO: We can remove this once the backend output format is more unified.
      if ('schema' in rawData) {
        return rawData as Data;
      }

      // This is a bson_table. We need to infer schema as the serialization
      // does not include the schema.
      // For now, `inferSchema` simply takes columns in first row and assume
      // they are 'object' type.
      const rows = rawData as TableRow[];
      return {
        schema: inferSchema(rows),
        data: rows,
      };
    }
  }
);

export const handleListIntegrationObjects = createAsyncThunk<
  string[],
  { apiKey: string; resourceId: string; forceLoad?: boolean },
  { state: RootState }
>(
  'resource/loadObjects',
  async (
    args: {
      apiKey: string;
      resourceId: string;
      forceLoad?: boolean;
    },
    thunkAPI
  ) => {
    // The resources are already defined, so just ignore this call if not force load.
    const state = thunkAPI.getState();
    const objects = state.resourceReducer.objectNames.names;
    if (objects && objects.length > 0 && !args.forceLoad) {
      return objects;
    }

    const { apiKey, resourceId } = args;
    const response = await fetch(
      `${apiAddress}/api/resource/${resourceId}/discover`,
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
  { apiKey: string; resourceId: string }
>(
  'resource/delete',
  async (
    args: {
      apiKey: string;
      resourceId: string;
      forceLoad?: boolean;
    },
    thunkAPI
  ) => {
    const { apiKey, resourceId } = args;
    const response = await fetch(
      `${apiAddress}/api/resource/${resourceId}/delete`,
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

export const handleTestConnectIntegration = createAsyncThunk<
  void,
  { apiKey: string; resourceId: string }
>(
  'resource/testConnect',
  async (
    args: {
      apiKey: string;
      resourceId: string;
    },
    thunkAPI
  ) => {
    const { apiKey, resourceId } = args;
    const response = await fetch(
      `${apiAddress}/api/resource/${resourceId}/test`,
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

export const handleConnectToNewIntegration = createAsyncThunk<
  void,
  {
    apiKey: string;
    service: Service;
    name: string;
    config: IntegrationConfig;
  }
>(
  'resource/connect',
  async (
    args: {
      apiKey: string;
      service: Service;
      name: string;
      config: IntegrationConfig;
    },
    thunkAPI
  ) => {
    const { apiKey, service, name, config } = args;
    Object.keys(config).forEach((k) => {
      if (config[k] === undefined) {
        config[k] = '';
      }
    });

    const res = await fetch(`${apiAddress}/api/resource/connect`, {
      method: 'POST',
      headers: {
        'api-key': apiKey,
        'resource-name': name,
        'resource-service': service,
        'resource-config': JSON.stringify(config),
      },
    });

    const responseBody = await res.json();

    if (!res.ok) {
      return thunkAPI.rejectWithValue(responseBody.error);
    }
  }
);

export const handleEditIntegration = createAsyncThunk<
  void,
  {
    apiKey: string;
    resourceId: string;
    name: string;
    config: IntegrationConfig;
  }
>(
  'resource/edit',
  async (
    args: {
      apiKey: string;
      resourceId: string;
      name: string;
      config: IntegrationConfig;
    },
    thunkAPI
  ) => {
    const { apiKey, resourceId, name, config } = args;

    Object.keys(config).forEach((k) => {
      if (!config[k]) {
        config[k] = '';
      }
    });

    const res = await fetch(
      `${apiAddress}/api/resource/${resourceId}/edit`,
      {
        method: 'POST',
        headers: {
          'api-key': apiKey,
          'resource-name': name,
          'resource-config': JSON.stringify(config),
        },
      }
    );

    const responseBody = await res.json();

    if (!res.ok) {
      return thunkAPI.rejectWithValue(responseBody.error);
    }
  }
);

export const resourceSlice = createSlice({
  name: 'resourceTablesReducer',
  initialState: initialState,
  reducers: {
    resetTestConnectStatus: (state) => {
      state.testConnectStatus = { loading: LoadingStatusEnum.Initial, err: '' };
    },
    resetConnectNewStatus: (state) => {
      state.connectNewStatus = { loading: LoadingStatusEnum.Initial, err: '' };
    },
    resetEditStatus: (state) => {
      state.editStatus = { loading: LoadingStatusEnum.Initial, err: '' };
    },
    resetDeletionStatus: (state) => {
      state.deletionStatus = { loading: LoadingStatusEnum.Initial, err: '' };
    },
  },
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
      state.testConnectStatus = { loading: LoadingStatusEnum.Loading, err: '' };
    });
    builder.addCase(handleTestConnectIntegration.fulfilled, (state) => {
      state.testConnectStatus = {
        loading: LoadingStatusEnum.Succeeded,
        err: '',
      };
    });
    builder.addCase(
      handleTestConnectIntegration.rejected,
      (state, { payload }) => {
        state.testConnectStatus = {
          loading: LoadingStatusEnum.Failed,
          err: payload as string,
        };
      }
    );
    builder.addCase(handleConnectToNewIntegration.pending, (state) => {
      state.connectNewStatus = { loading: LoadingStatusEnum.Loading, err: '' };
    });
    builder.addCase(handleConnectToNewIntegration.fulfilled, (state) => {
      state.connectNewStatus = {
        loading: LoadingStatusEnum.Succeeded,
        err: '',
      };
    });
    builder.addCase(
      handleConnectToNewIntegration.rejected,
      (state, { payload }) => {
        state.connectNewStatus = {
          loading: LoadingStatusEnum.Failed,
          err: payload as string,
        };
      }
    );
    builder.addCase(handleEditIntegration.pending, (state) => {
      state.editStatus = { loading: LoadingStatusEnum.Loading, err: '' };
    });
    builder.addCase(handleEditIntegration.fulfilled, (state) => {
      state.editStatus = {
        loading: LoadingStatusEnum.Succeeded,
        err: '',
      };
    });
    builder.addCase(handleEditIntegration.rejected, (state, { payload }) => {
      state.editStatus = {
        loading: LoadingStatusEnum.Failed,
        err: payload as string,
      };
    });
  },
});

export const {
  resetTestConnectStatus,
  resetConnectNewStatus,
  resetDeletionStatus,
  resetEditStatus,
} = resourceSlice.actions;

export default resourceSlice.reducer;
