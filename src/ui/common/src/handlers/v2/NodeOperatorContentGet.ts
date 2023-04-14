// This file should map exactly to
// src/golang/cmd/server/handler/v2/node_operator_content_get.go

import { APIKeyParameter } from '../parameters/Header';
import {
  DagIdParameter,
  NodeIdParameter,
  WorkflowIdParameter,
} from '../parameters/Path';
import { NodeContentResponse } from '../responses/Node';

export type NodeOperatorContentGetRequest = APIKeyParameter &
  DagIdParameter &
  NodeIdParameter &
  WorkflowIdParameter;
export type NodeOperatorContentGetResponse = NodeContentResponse;

export const nodeOperatorContentGetQuery = (
  req: NodeOperatorContentGetRequest
) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}/node/operator/${req.nodeId}/content`,
  headers: { 'api-key': req.apiKey },
});
