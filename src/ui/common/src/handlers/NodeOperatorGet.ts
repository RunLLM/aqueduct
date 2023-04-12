// This file should map exactly to
// src/golang/cmd/server/handler/v2/dag_get.go

import { APIKeyRequest } from './requests/ApiKey';
import { OperatorResponse } from './responses/node';

export type NodeOperatorGetRequest = APIKeyRequest & {
  workflowId: string;
  dagId: string;
  nodeId: string;
};

export type NodeOperatorGetResponse = OperatorResponse;

export const nodeOperatorGetQuery = (req: NodeOperatorGetRequest) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}/node/operator/${req.nodeId}`,
  headers: { 'api-key': req.apiKey },
});
