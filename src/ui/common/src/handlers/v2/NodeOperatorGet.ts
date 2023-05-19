// This file should map exactly to
// src/golang/cmd/server/handler/v2/node_operator_get.go

import { APIKeyParameter } from '../parameters/Header';
import {
  DagIdParameter,
  NodeIdParameter,
  WorkflowIdParameter,
} from '../parameters/Path';
import { OperatorResponse } from '../responses/node';

export type NodeOperatorGetRequest = APIKeyParameter &
  DagIdParameter &
  NodeIdParameter &
  WorkflowIdParameter;
export type NodeOperatorGetResponse = OperatorResponse;

export const nodeOperatorGetQuery = (req: NodeOperatorGetRequest) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}/node/operator/${req.nodeId}`,
  headers: { 'api-key': req.apiKey },
});
