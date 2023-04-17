// This file should map exactly to
// src/golang/cmd/server/handler/v2/workflows_get.go

import { APIKeyParameter } from '../parameters/Header';
import { WorkflowResponse } from '../responses/Workflow';

export type WorkflowsGetRequest = APIKeyParameter;

export type WorkflowsGetResponse = WorkflowResponse[];

export const workflowsGetQuery = (req: WorkflowsGetRequest) => ({
  url: `workflows/`,
  headers: { 'api-key': req.apiKey },
});
