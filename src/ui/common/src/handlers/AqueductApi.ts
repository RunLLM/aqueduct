import * as rtkQueryRaw from '@reduxjs/toolkit/dist/query/react/index.js';
import { FetchBaseQueryError } from '@reduxjs/toolkit/query/react';

import { apiAddress } from '../components/hooks/useAqueductConsts';
import {
  storageMigrationListQuery,
  storageMigrationListRequest,
  storageMigrationListResponse,
} from './ListStorageMigrations';
import {
  workflowGetQuery,
  WorkflowGetRequest,
  WorkflowGetResponse,
} from './WorkflowGet';

const { createApi, fetchBaseQuery } = ((rtkQueryRaw as any).default ??
  rtkQueryRaw) as typeof rtkQueryRaw;

const transformErrorResponse = (resp: FetchBaseQueryError) =>
  (resp.data as { error: string })?.error;

export const aqueductApi = createApi({
  reducerPath: 'aqueductApi',
  baseQuery: fetchBaseQuery({ baseUrl: `${apiAddress}/api/v2/` }),
  keepUnusedDataFor: 60,
  endpoints: (builder) => ({
    workflowGet: builder.query<WorkflowGetResponse, WorkflowGetRequest>({
      query: (req) => workflowGetQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
    storageMigrationList: builder.query<
      storageMigrationListResponse,
      storageMigrationListRequest
    >({
      query: (req) => storageMigrationListQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
  }),
});

export const { useWorkflowGetQuery, useStorageMigrationListQuery } =
  aqueductApi;
