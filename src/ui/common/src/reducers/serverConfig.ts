import { createSlice } from '@reduxjs/toolkit';

import { handleGetServerConfig } from '../handlers/getServerConfig';
import { LoadingStatus, LoadingStatusEnum } from '../utils/shared';

export type ServerConfig = {
  aqPath: string;
  retentionJobPeriod: string;
  apiKey: string;
  storageConfig: {
    type: string;
    fileConfig?: {
      directory: string;
    };
    gcsConfig?: {
      bucket: string;
    };
    s3Config?: {
      region: string;
      bucket: string;
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

    builder.addCase(
      handleGetServerConfig.fulfilled,
      (state, { meta, payload }) => {
        state.status = { loading: LoadingStatusEnum.Succeeded, err: '' };
        state.config = payload as ServerConfig;
      }
    );

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
