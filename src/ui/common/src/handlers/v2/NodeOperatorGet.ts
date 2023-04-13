// This file should map exactly to
// src/golang/cmd/server/handler/v2/node_operator_get.go

import { APIKeyParameter } from '../parameters/ApiKey';
import { DagIdParameter } from '../parameters/DagId';
import { NodeIdParameter } from '../parameters/NodeId';
import { WorkflowIdParameter } from '../parameters/WorkflowId';
import { OperatorResponse } from '../responses/Node';

export type NodeOperatorGetRequest = APIKeyParameter &
  DagIdParameter &
  NodeIdParameter &
  WorkflowIdParameter;
export type NodeOperatorGetResponse = OperatorResponse;

export const nodeOperatorGetQuery = (req: NodeOperatorGetRequest) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}/node/operator/${req.nodeId}`,
  headers: { 'api-key': req.apiKey },
});
