// This file should map exactly to
// src/golang/cmd/server/handler/v2/resource_workflows_get.go

import { APIKeyParameter } from '../parameters/Header';
import { ResourceIdParameter } from '../parameters/Path';
import { WorkflowAndDagIDResponse } from '../responses/workflow';

export type ResourceWorkflowsGetRequest = APIKeyParameter &
  ResourceIdParameter;

// A list of workflow IDs that use this resource.
export type ResourceWorkflowsGetResponse = WorkflowAndDagIDResponse[];

export const resourceWorkflowsGetQuery = (
  req: ResourceWorkflowsGetRequest
) => ({
  url: `resource/${req.resourceId}/workflows`,
  headers: { 'api-key': req.apiKey },
});
