// This file should map exactly to
// src/golang/cmd/server/handler/v2/node_check_get.go

import { APIKeyParameter } from '../parameters/Header';
import {
  DagIdParameter,
  NodeIdParameter,
  WorkflowIdParameter,
} from '../parameters/Path';
import { MergedNodeResponse } from '../responses/Node';

export type NodeCheckGetRequest = APIKeyParameter &
  DagIdParameter &
  NodeIdParameter &
  WorkflowIdParameter;

export type NodeCheckGetResponse = MergedNodeResponse;

export const nodeMetricGetQuery = (req: NodeCheckGetRequest) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}/node/check/${req.nodeId}`,
  headers: { 'api-key': req.apiKey },
});
