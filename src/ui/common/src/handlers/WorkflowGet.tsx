// This file should map exactly to
// src/golang/cmd/server/handler/v2/workflow_get.go

import { APIKeyRequest } from './requests/ApiKey';
import { WorkflowResponse } from './responses/workflow';

export type WorkflowGetRequest = APIKeyRequest & {
  workflowId: string;
};

export type WorkflowGetResponse = WorkflowResponse;

export const workflowGetQuery = (req: WorkflowGetRequest) => ({
  url: `workflow/${req.workflowId}`,
  headers: { 'api-key': req.apiKey },
});
