import {
  createApi,
  fetchBaseQuery,
  FetchBaseQueryError,
} from '@reduxjs/toolkit/query/react';

import { apiAddress } from '../components/hooks/useAqueductConsts';
import {
  workflowGetQuery,
  WorkflowGetRequest,
  WorkflowGetResponse,
} from './WorkflowGet';
import {
  storageMigrationListQuery,
  storageMigrationListRequest,
  storageMigrationListResponse
} from "./ListStorageMigrations";

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
    storageMigrationList: builder.query<storageMigrationListResponse, storageMigrationListRequest>({
      query: (req) => storageMigrationListQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
  }),
});

export const { useWorkflowGetQuery, useStorageMigrationListQuery } = aqueductApi;
