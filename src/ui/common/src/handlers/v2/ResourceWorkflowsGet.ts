// This file should map exactly to
// src/golang/cmd/server/handler/v2/resource_workflows_get.go

import { APIKeyParameter } from '../parameters/Header';
import { IntegrationIdParameter } from '../parameters/Path';
import { WorkflowAndDagIDResponse } from '../responses/workflow';

export type IntegrationWorkflowsGetRequest = APIKeyParameter &
  IntegrationIdParameter;

// A list of workflow IDs that use this integration.
export type IntegrationWorkflowsGetResponse = WorkflowAndDagIDResponse[];

export const integrationWorkflowsGetQuery = (
  req: IntegrationWorkflowsGetRequest
) => ({
  url: `resource/${req.integrationId}/workflows`,
  headers: { 'api-key': req.apiKey },
});
