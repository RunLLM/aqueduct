// This file should map exactly to
// src/golang/cmd/server/handler/v2/node_metric_results_get.go

import { APIKeyParameter } from '../parameters/Header';
import {
  DagIdParameter,
  NodeIdParameter,
  WorkflowIdParameter,
} from '../parameters/Path';
import { OperatorWithArtifactNodeResultResponse } from '../responses/node';

export type NodeMetricResultsGetRequest = APIKeyParameter &
  DagIdParameter &
  NodeIdParameter &
  WorkflowIdParameter;

export type NodeMetricResultsGetResponse = OperatorWithArtifactNodeResultResponse[];

export const nodeMetricResultsGetQuery = (
  req: NodeMetricResultsGetResponse
) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}/node/metric/${req.nodeId}/results`,
  headers: { 'api-key': req.apiKey },
});
