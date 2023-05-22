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
import { dagsGetQuery, DagsGetRequest, DagsGetResponse } from './v2/DagsGet';
import {
  environmentGetQuery,
  EnvironmentGetRequest,
  EnvironmentGetResponse,
} from './v2/EnvironmentGet';
import {
  integrationOperatorsGetQuery,
  IntegrationOperatorsGetRequest,
  IntegrationOperatorsGetResponse,
} from './v2/IntegrationOperatorsGet';
import {
  integrationsWorkflowsGetQuery,
  IntegrationsWorkflowsGetRequest,
  IntegrationsWorkflowsGetResponse,
} from './v2/IntegrationsWorkflowsGet';
import {
  integrationWorkflowsGetQuery,
  IntegrationWorkflowsGetRequest,
  IntegrationWorkflowsGetResponse,
} from './v2/IntegrationWorkflowsGet';
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
  nodeCheckGetQuery,
  NodeCheckGetRequest,
  NodeCheckGetResponse,
} from './v2/NodeCheckGet';
import {
  nodeCheckResultContentGetQuery,
  NodeCheckResultContentGetRequest,
  NodeCheckResultContentGetResponse,
} from './v2/NodeCheckResultContentGet';
import {
  nodeMetricGetQuery,
  NodeMetricGetRequest,
  NodeMetricGetResponse,
} from './v2/NodeMetricGet';
import {
  nodeMetricResultContentGetQuery,
  NodeMetricResultContentGetRequest,
  NodeMetricResultContentGetResponse,
} from './v2/NodeMetricResultContentGet';
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
  workflowDeletePostQuery,
  WorkflowDeletePostRequest,
  WorkflowDeletePostResponse,
} from './v2/WorkflowDeletePost';
import {
  workflowEditPostQuery,
  WorkflowEditPostRequest,
  WorkflowEditPostResponse,
} from './v2/WorkflowEditPost';
import {
  workflowGetQuery,
  WorkflowGetRequest,
  WorkflowGetResponse,
} from './v2/WorkflowGet';
import {
  workflowObjectsGetQuery,
  WorkflowObjectsGetRequest,
  WorkflowObjectsGetResponse,
} from './v2/WorkflowObjectsGet';
import {
  workflowsGetQuery,
  WorkflowsGetRequest,
  WorkflowsGetResponse,
} from './v2/WorkflowsGet';
import {
  workflowTriggerPostQuery,
  WorkflowTriggerPostRequest,
  WorkflowTriggerPostResponse,
} from './v2/WorkflowTriggerPost';

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
    dagsGet: builder.query<DagsGetResponse, DagsGetRequest>({
      query: (req) => dagsGetQuery(req),
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
    environmentGet: builder.query<
      EnvironmentGetResponse,
      EnvironmentGetRequest
    >({
      query: (req) => environmentGetQuery(req),
      transformErrorResponse,
    }),
    integrationOperatorsGet: builder.query<
      IntegrationOperatorsGetResponse,
      IntegrationOperatorsGetRequest
    >({
      query: (req) => integrationOperatorsGetQuery(req),
      transformErrorResponse,
    }),
    integrationWorkflowsGet: builder.query<
      IntegrationWorkflowsGetResponse,
      IntegrationWorkflowsGetRequest
    >({
      query: (req) => integrationWorkflowsGetQuery(req),
      transformErrorResponse,
    }),
    integrationsWorkflowsGet: builder.query<
      IntegrationsWorkflowsGetResponse,
      IntegrationsWorkflowsGetRequest
    >({
      query: (req) => integrationsWorkflowsGetQuery(req),
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
    nodeMetricGet: builder.query<NodeMetricGetResponse, NodeMetricGetRequest>({
      query: (req) => nodeMetricGetQuery(req),
      transformErrorResponse,
    }),
    nodeMetricResultContentGet: builder.query<
      NodeMetricResultContentGetResponse,
      NodeMetricResultContentGetRequest
    >({
      query: (req) => nodeMetricResultContentGetQuery(req),
      transformErrorResponse,
    }),
    nodeCheckGet: builder.query<NodeCheckGetResponse, NodeCheckGetRequest>({
      query: (req) => nodeCheckGetQuery(req),
      transformErrorResponse,
    }),
    nodeCheckResultContentGet: builder.query<
      NodeCheckResultContentGetResponse,
      NodeCheckResultContentGetRequest
    >({
      query: (req) => nodeCheckResultContentGetQuery(req),
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
    workflowDeletePost: builder.mutation<
      WorkflowDeletePostResponse,
      WorkflowDeletePostRequest
    >({
      query: (req) => workflowDeletePostQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
    workflowEditPost: builder.mutation<
      WorkflowEditPostResponse,
      WorkflowEditPostRequest
    >({
      query: (req) => workflowEditPostQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
    workflowTriggerPost: builder.mutation<
      WorkflowTriggerPostResponse,
      WorkflowTriggerPostRequest
    >({
      query: (req) => workflowTriggerPostQuery(req),
      transformErrorResponse: transformErrorResponse,
    }),
    workflowObjectsGet: builder.query<
      WorkflowObjectsGetResponse,
      WorkflowObjectsGetRequest
    >({
      query: (req) => workflowObjectsGetQuery(req),
      transformErrorResponse,
    }),
    workflowsGet: builder.query<WorkflowsGetResponse, WorkflowsGetRequest>({
      query: (req) => workflowsGetQuery(req),
      transformErrorResponse,
    }),
    workflowGet: builder.query<WorkflowGetResponse, WorkflowGetRequest>({
      query: (req) => workflowGetQuery(req),
      transformErrorResponse,
    }),
  }),
});

export const {
  useDagGetQuery,
  useDagsGetQuery,
  useDagOperatorsGetQuery,
  useDagResultGetQuery,
  useDagResultsGetQuery,
  useEnvironmentGetQuery,
  useIntegrationOperatorsGetQuery,
  useIntegrationWorkflowsGetQuery,
  useIntegrationsWorkflowsGetQuery,
  useStorageMigrationListQuery,
  useNodeArtifactGetQuery,
  useNodeArtifactResultContentGetQuery,
  useNodeArtifactResultsGetQuery,
  useNodeOperatorGetQuery,
  useNodeOperatorContentGetQuery,
  useNodeMetricGetQuery,
  useNodeMetricResultContentGetQuery,
  useNodeCheckGetQuery,
  useNodeCheckResultContentGetQuery,
  useNodesGetQuery,
  useNodesResultsGetQuery,
  useWorkflowGetQuery,
  useWorkflowObjectsGetQuery,
  useWorkflowsGetQuery,
  useWorkflowDeletePostMutation,
  useWorkflowEditPostMutation,
  useWorkflowTriggerPostMutation,
} = aqueductApi;
