import { createAsyncThunk } from '@reduxjs/toolkit';

import { apiAddress } from '../components/hooks/useAqueductConsts';
import { ServerConfig } from '../reducers/serverConfig';

export const handleGetServerConfig = createAsyncThunk<
  ServerConfig,
  {
    apiKey: string;
  }
>(
  'api/get_server_config',
  async (
    args: {
      apiKey: string;
    },
    thunkAPI
  ) => {
    const { apiKey } = args;
    const res = await fetch(`${apiAddress}/api/config`, {
      method: 'GET',
      headers: {
        'api-key': apiKey,
      },
    });

    const body = await res.json();

    if (!res.ok) {
      return thunkAPI.rejectWithValue(body.error);
    }

    return body as ServerConfig;
  }
);
