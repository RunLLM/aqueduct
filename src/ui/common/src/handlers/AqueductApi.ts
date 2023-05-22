import * as rtkQueryRaw from '@reduxjs/toolkit/dist/query/react/index.js';
import { FetchBaseQueryError } from '@reduxjs/toolkit/query/react';

import { apiAddress } from '../components/hooks/useAqueductConsts';
import { dagGetQuery, DagGetRequest, DagGetResponse } from './v2/DagGet';
import {
  dagOperatorsGetQuery,
  DagOperatorsGetRequest,
  DagOperatorsGetResponse,
} from './v2/DagOperatorsGet';
import {
  dagResultGetQuery,
  DagResultGetRequest,
  DagResultGetResponse,
} from './v2/DagResultGet';
import {
  dagResultsGetQuery,
  DagResultsGetRequest,
  DagResultsGetResponse,
} from './v2/DagResultsGet';
import {
  resourceOperatorsGetQuery,
  ResourceOperatorsGetRequest,
  ResourceOperatorsGetResponse,
} from './v2/ResourceOperatorsGet';
import {
  resourcesWorkflowsGetQuery,
  ResourcesWorkflowsGetRequest,
  ResourcesWorkflowsGetResponse,
} from './v2/ResourcesWorkflowsGet';
import {
  resourceWorkflowsGetQuery,
  ResourceWorkflowsGetRequest,
  ResourceWorkflowsGetResponse,
} from './v2/ResourceWorkflowsGet';
import {
  storageMigrationListQuery,
  storageMigrationListRequest,
  storageMigrationListResponse,
} from './v2/ListStorageMigrations';
import {
  nodeArtifactGetQuery,
  NodeArtifactGetRequest,
  NodeArtifactGetResponse,
} from './v2/NodeArtifactGet';
import {
  nodeArtifactResultContentGetQuery,
  NodeArtifactResultContentGetRequest,
  NodeArtifactResultContentGetResponse,
} from './v2/NodeArtifactResultContentGet';
import {
  nodeArtifactResultsGetQuery,
  NodeArtifactResultsGetRequest,
  NodeArtifactResultsGetResponse,
} from './v2/NodeArtifactResultsGet';
import {
  nodeOperatorContentGetQuery,
  NodeOperatorContentGetRequest,
  NodeOperatorContentGetResponse,
} from './v2/NodeOperatorContentGet';
import {
  nodeOperatorGetQuery,
  NodeOperatorGetRequest,
  NodeOperatorGetResponse,
} from './v2/NodeOperatorGet';
import {
  nodesGetQuery,
  NodesGetRequest,
  NodesGetResponse,
} from './v2/NodesGet';
import {
  nodesResultsGetQuery,
  NodesResultsGetRequest,
  NodesResultsGetResponse,
} from './v2/NodesResultsGet';
import {
  workflowGetQuery,
  WorkflowGetRequest,
  WorkflowGetResponse,
} from './v2/WorkflowGet';
import {
  workflowsGetQuery,
  WorkflowsGetRequest,
  WorkflowsGetResponse,
} from './v2/WorkflowsGet';

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
      transformErrorResponse,
    }),
    dagOperatorsGet: builder.query<
      DagOperatorsGetResponse,
      DagOperatorsGetRequest
    >({
      query: (req) => dagOperatorsGetQuery(req),
      transformErrorResponse,
    }),
    dagResultGet: builder.query<DagResultGetResponse, DagResultGetRequest>({
      query: (req) => dagResultGetQuery(req),
      transformErrorResponse,
    }),
    dagResultsGet: builder.query<DagResultsGetResponse, DagResultsGetRequest>({
      query: (req) => dagResultsGetQuery(req),
      transformErrorResponse,
    }),
    resourceOperatorsGet: builder.query<
      ResourceOperatorsGetResponse,
      ResourceOperatorsGetRequest
    >({
      query: (req) => resourceOperatorsGetQuery(req),
      transformErrorResponse,
    }),
    resourceWorkflowsGet: builder.query<
      ResourceWorkflowsGetResponse,
      ResourceWorkflowsGetRequest
    >({
      query: (req) => resourceWorkflowsGetQuery(req),
      transformErrorResponse,
    }),
    resourcesWorkflowsGet: builder.query<
      ResourcesWorkflowsGetResponse,
      ResourcesWorkflowsGetRequest
    >({
      query: (req) => resourcesWorkflowsGetQuery(req),
      transformErrorResponse,
    }),
    nodeArtifactGet: builder.query<
      NodeArtifactGetResponse,
      NodeArtifactGetRequest
    >({
      query: (req) => nodeArtifactGetQuery(req),
      transformErrorResponse,
    }),
    nodeArtifactResultContentGet: builder.query<
      NodeArtifactResultContentGetResponse,
      NodeArtifactResultContentGetRequest
    >({
      query: (req) => nodeArtifactResultContentGetQuery(req),
      transformErrorResponse,
    }),
    nodeArtifactResultsGet: builder.query<
      NodeArtifactResultsGetResponse,
      NodeArtifactResultsGetRequest
    >({
      query: (req) => nodeArtifactResultsGetQuery(req),
      transformErrorResponse,
    }),
    nodeOperatorGet: builder.query<
      NodeOperatorGetResponse,
      NodeOperatorGetRequest
    >({
      query: (req) => nodeOperatorGetQuery(req),
      transformErrorResponse,
    }),
    nodeOperatorContentGet: builder.query<
      NodeOperatorContentGetResponse,
      NodeOperatorContentGetRequest
    >({
      query: (req) => nodeOperatorContentGetQuery(req),
      transformErrorResponse,
    }),
    nodesGet: builder.query<NodesGetResponse, NodesGetRequest>({
      query: (req) => nodesGetQuery(req),
      transformErrorResponse,
    }),
    nodesResultsGet: builder.query<
      NodesResultsGetResponse,
      NodesResultsGetRequest
    >({
      query: (req) => nodesResultsGetQuery(req),
      transformErrorResponse,
    }),
    storageMigrationList: builder.query<
      storageMigrationListResponse,
      storageMigrationListRequest
    >({
      query: (req) => storageMigrationListQuery(req),
      transformErrorResponse,
    }),
    workflowsGet: builder.query<WorkflowsGetResponse, WorkflowsGetRequest>({
      query: (req) => workflowsGetQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
    workflowGet: builder.query<WorkflowGetResponse, WorkflowGetRequest>({
      query: (req) => workflowGetQuery(req),
      transformErrorResponse,
    }),
  }),
});

export const {
  useDagGetQuery,
  useDagOperatorsGetQuery,
  useDagResultGetQuery,
  useDagResultsGetQuery,
  useResourceOperatorsGetQuery,
  useResourceWorkflowsGetQuery,
  useResourcesWorkflowsGetQuery,
  useStorageMigrationListQuery,
  useNodeArtifactGetQuery,
  useNodeArtifactResultContentGetQuery,
  useNodeArtifactResultsGetQuery,
  useNodeOperatorGetQuery,
  useNodeOperatorContentGetQuery,
  useNodesGetQuery,
  useNodesResultsGetQuery,
  useWorkflowGetQuery,
  useWorkflowsGetQuery,
} = aqueductApi;
