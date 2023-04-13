// This file should map exactly to
// src/golang/cmd/server/handler/v2/workflow_get.go
import { APIKeyParameter } from '../parameters/ApiKey';
import { WorkflowIdParameter } from '../parameters/WorkflowId';
import { WorkflowResponse } from '../responses/Workflow';

export type WorkflowGetRequest = APIKeyParameter & WorkflowIdParameter;

export type WorkflowGetResponse = WorkflowResponse;

export const workflowGetQuery = (req: WorkflowGetRequest) => ({
  url: `workflow/${req.workflowId}`,
  headers: { 'api-key': req.apiKey },
});
