import { createSlice } from '@reduxjs/toolkit';

import { handleGetServerConfig } from '../handlers/getServerConfig';
import { ExecState, LoadingStatus, LoadingStatusEnum } from '../utils/shared';

export type ServerConfig = {
  aqPath: string;
  retentionJobPeriod: string;
  apiKey: string;
  storageConfig: {
    type: string;
    integration_id?: string;
    integration_name: string;
    connected_at?: number;
    exec_state: ExecState;
    fileConfig?: {
      directory: string;
    };
    gcsConfig?: {
      bucket: string;
    };
    s3Config?: {
      region: string;
      bucket: string;

      // If set, expected to be in the format `path/to/dir/`
      root_dir: string;
    };
  };
};

// Create a config object here and
export type ServerConfigState = {
  status: LoadingStatus;
  config?: ServerConfig;
};

const initialState: ServerConfigState = {
  status: { loading: LoadingStatusEnum.Initial, err: '' },
  config: null,
};

export const serverConfigSlice = createSlice({
  name: 'serverConfigReducer',
  initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder.addCase(handleGetServerConfig.pending, (state) => {
      state.status = { loading: LoadingStatusEnum.Loading, err: '' };
    });

    builder.addCase(handleGetServerConfig.fulfilled, (state, { payload }) => {
      state.status = { loading: LoadingStatusEnum.Succeeded, err: '' };
      state.config = payload as ServerConfig;
    });

    builder.addCase(handleGetServerConfig.rejected, (state, { payload }) => {
      state.config = null;
      state.status = {
        loading: LoadingStatusEnum.Failed,
        err: payload as string,
      };
    });
  },
});

export default serverConfigSlice.reducer;
