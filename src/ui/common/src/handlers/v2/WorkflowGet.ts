// This file should map exactly to
// src/golang/cmd/server/handler/v2/workflow_get.go
import { APIKeyParameter } from '../parameters/Header';
import { WorkflowIdParameter } from '../parameters/Path';
import { WorkflowResponse } from '../responses/workflow';

export type WorkflowGetRequest = APIKeyParameter & WorkflowIdParameter;

export type WorkflowGetResponse = WorkflowResponse;

export const workflowGetQuery = (req: WorkflowGetRequest) => ({
  url: `workflow/${req.workflowId}`,
  headers: { 'api-key': req.apiKey },
});
