// This file should map exactly to
// src/golang/cmd/server/handler/v2/dag_get.go

import { APIKeyRequest } from './requests/ApiKey';
import { DagResponse } from './responses/workflow';

export type DagGetRequest = APIKeyRequest & {
  workflowId: string;
  dagId: string;
};

export type DagGetResponse = DagResponse;

export const dagGetQuery = (req: DagGetRequest) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}`,
  headers: { 'api-key': req.apiKey },
});
