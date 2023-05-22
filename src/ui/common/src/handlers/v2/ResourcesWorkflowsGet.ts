// This file should map exactly to
// src/golang/cmd/server/handler/v2/resources_workflows_get.go

import { APIKeyParameter } from '../parameters/Header';
import { WorkflowAndDagIDResponse } from '../responses/workflow';

export type ResourcesWorkflowsGetRequest = APIKeyParameter;

// ResourceID -> list of workflows that use this resource.
export type ResourcesWorkflowsGetResponse = {
  [resourceID: string]: WorkflowAndDagIDResponse[];
};

export const resourcesWorkflowsGetQuery = (
  req: ResourcesWorkflowsGetRequest
) => ({
  url: `resources/workflows`,
  headers: { 'api-key': req.apiKey },
});
