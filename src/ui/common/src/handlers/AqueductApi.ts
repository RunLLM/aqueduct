import {
  createApi,
  fetchBaseQuery,
  FetchBaseQueryError,
} from '@reduxjs/toolkit/query/react';

import { apiAddress } from '../components/hooks/useAqueductConsts';
import { getDagQuery, GetDagRequest, GetDagResponse } from './GetDag';
import {
  getDagResultQuery,
  GetDagResultRequest,
  GetDagResultResponse,
} from './GetDagResult';

const transformErrorResponse = (resp: FetchBaseQueryError) =>
  (resp.data as { error: string })?.error;

export const aqueductApi = createApi({
  reducerPath: 'aqueductApi',
  baseQuery: fetchBaseQuery({ baseUrl: `${apiAddress}/api/` }),
  keepUnusedDataFor: 60,
  endpoints: (builder) => ({
    getDag: builder.query<GetDagResponse, GetDagRequest>({
      query: (req) => getDagQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
    getDagResult: builder.query<GetDagResultResponse, GetDagResultRequest>({
      query: (req) => getDagResultQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
  }),
});

export const { useGetDagQuery, useGetDagResultQuery } = aqueductApi;
