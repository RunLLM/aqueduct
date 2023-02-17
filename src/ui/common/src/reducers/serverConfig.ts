import { createSlice } from '@reduxjs/toolkit';

import { handleGetServerConfig } from '../handlers/getServerConfig';
import { LoadingStatus, LoadingStatusEnum } from '../utils/shared';

type ServerConfig = {
  aqPath: string;
  encryptionKey: string;
  retentionJobPeriod: string;
  apiKey: string;
  storageConfig: {
    type: string;
    file_config?: {
      directory: string;
    };
    gcs_config?: {
      bucket: string;
      service_account_credentials: string;
    };
    s3_config?: {
      region: string;
      bucket: string;
      credentials_path: string;
      credentials_profile: string;
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
      state = { status: { loading: LoadingStatusEnum.Loading, err: '' } };
    });

    builder.addCase(
      handleGetServerConfig.fulfilled,
      (state, { meta, payload }) => {
        state = {
          status: { loading: LoadingStatusEnum.Succeeded, err: '' },
          config: payload as ServerConfig,
        };
      }
    );

    builder.addCase(handleGetServerConfig.rejected, (state, { payload }) => {
      state = {
        status: { loading: LoadingStatusEnum.Failed, err: payload as string },
        config: null,
      };
    });
  },
});

export default serverConfigSlice.reducer;
