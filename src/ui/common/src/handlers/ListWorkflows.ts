// This file should map exactly to
// src/golang/cmd/server/handler/v2/list_workflows.go

import { APIKeyRequest } from './requests/ApiKey';
import { WorkflowResponse } from './responses/workflow';

export type workflowListRequest = APIKeyRequest

export type workflowListResponse = WorkflowResponse[];

export const workflowListQuery = (
  req: workflowListRequest
) => ({
  url: 'workflows',
  headers: {
    'api-key': req.apiKey,
  },
});
