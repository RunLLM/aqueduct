// This file should map exactly to
// src/golang/cmd/server/handler/v2/dag_get.go

import { APIKeyRequest } from './requests/ApiKey';
import { DagResultResponse } from './responses/workflow';

export type DagResultGetRequest = APIKeyRequest & {
  workflowId: string;
  dagResultId: string;
};

export type DagResultGetResponse = DagResultResponse;

export const dagResultGetQuery = (req: DagResultGetRequest) => ({
  url: `workflow/${req.workflowId}/result/${req.dagResultId}`,
  headers: { 'api-key': req.apiKey },
});
