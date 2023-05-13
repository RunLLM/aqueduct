// This file should map exactly to
// src/golang/cmd/server/handler/v2/node_metric_get.go

import { APIKeyParameter } from '../parameters/Header';
import {
  DagIdParameter,
  NodeIdParameter,
  WorkflowIdParameter,
} from '../parameters/Path';
import { MergedNodeResponse } from '../responses/Node';

export type NodeMetricGetRequest = APIKeyParameter &
  DagIdParameter &
  NodeIdParameter &
  WorkflowIdParameter;

export type NodeMetricGetResponse = MergedNodeResponse;

export const nodeMetricGetQuery = (req: NodeMetricGetRequest) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}/node/metric/${req.nodeId}`,
  headers: { 'api-key': req.apiKey },
});
