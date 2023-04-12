import * as rtkQueryRaw from '@reduxjs/toolkit/dist/query/react/index.js';
import { FetchBaseQueryError } from '@reduxjs/toolkit/query/react';

import { apiAddress } from '../components/hooks/useAqueductConsts';
import { dagGetQuery, DagGetRequest, DagGetResponse } from './DagGet';
import {
  dagResultGetQuery,
  DagResultGetRequest,
  DagResultGetResponse,
} from './DagResultGet';
import {
  storageMigrationListQuery,
  storageMigrationListRequest,
  storageMigrationListResponse,
} from './ListStorageMigrations';
import {
  nodeArtifactGetQuery,
  NodeArtifactGetRequest,
  NodeArtifactGetResponse,
} from './NodeArtifactGet';
import {
  nodeOperatorGetQuery,
  NodeOperatorGetRequest,
  NodeOperatorGetResponse,
} from './NodeOperatorGet';
import { nodesGetQuery, NodesGetRequest, NodesGetResponse } from './NodesGet';
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
    dagGet: builder.query<DagGetResponse, DagGetRequest>({
      query: (req) => dagGetQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
    dagResultGet: builder.query<DagResultGetResponse, DagResultGetRequest>({
      query: (req) => dagResultGetQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
    nodeOperatorGet: builder.query<
      NodeOperatorGetResponse,
      NodeOperatorGetRequest
    >({
      query: (req) => nodeOperatorGetQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
    nodeArtifactGet: builder.query<
      NodeArtifactGetResponse,
      NodeArtifactGetRequest
    >({
      query: (req) => nodeArtifactGetQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
    nodesGet: builder.query<NodesGetResponse, NodesGetRequest>({
      query: (req) => nodesGetQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
    storageMigrationList: builder.query<
      storageMigrationListResponse,
      storageMigrationListRequest
    >({
      query: (req) => storageMigrationListQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
    workflowGet: builder.query<WorkflowGetResponse, WorkflowGetRequest>({
      query: (req) => workflowGetQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
  }),
});

export const {
  useDagGetQuery,
  useDagResultGetQuery,
  useStorageMigrationListQuery,
  useNodeArtifactGetQuery,
  useNodeOperatorGetQuery,
  useNodesGetQuery,
  useWorkflowGetQuery,
} = aqueductApi;
