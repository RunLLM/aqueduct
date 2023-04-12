// This file should map exactly to
// src/golang/cmd/server/handler/v2/dag_get.go

import { APIKeyRequest } from './requests/ApiKey';
import { NodesResponse } from './responses/node';

export type NodesGetRequest = APIKeyRequest & {
  workflowId: string;
  dagId: string;
};

export type NodesGetResponse = NodesResponse;

export const nodesGetQuery = (req: NodesGetRequest) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}/nodes`,
  headers: { 'api-key': req.apiKey },
});
