import {
  createApi,
  fetchBaseQuery,
  FetchBaseQueryError,
} from '@reduxjs/toolkit/query/react';

import { apiAddress } from '../components/hooks/useAqueductConsts';
import { dagGetQuery, DagGetRequest, DagGetResponse } from './DagGet';
import {
  dagResultGetQuery,
  DagResultGetRequest,
  DagResultGetResponse,
} from './DagResultGet';
import {
  workflowGetQuery,
  WorkflowGetRequest,
  WorkflowGetResponse,
} from './WorkflowGet';

const transformErrorResponse = (resp: FetchBaseQueryError) =>
  (resp.data as { error: string })?.error;

export const aqueductApi = createApi({
  reducerPath: 'aqueductApi',
  baseQuery: fetchBaseQuery({ baseUrl: `${apiAddress}/api/v2/` }),
  keepUnusedDataFor: 60,
  endpoints: (builder) => ({
    dagGet: builder.query<DagGetResponse, DagGetRequest>({
      query: (req) => dagGetQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
    dagResultGet: builder.query<DagResultGetResponse, DagResultGetRequest>({
      query: (req) => dagResultGetQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
    workflowGet: builder.query<WorkflowGetResponse, WorkflowGetRequest>({
      query: (req) => workflowGetQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
  }),
});

export const { useDagGetQuery, useDagResultGetQuery, useWorkflowGetQuery } =
  aqueductApi;