// This file should map exactly to
// src/golang/cmd/server/handler/v2/integration_workflows.go

import { APIKeyParameter } from '../parameters/Header';
import { IntegrationIdParameter } from '../parameters/Path';

export type IntegrationWorkflowsGetRequest = APIKeyParameter &
  IntegrationIdParameter;

// A list of workflow IDs that use this integration.
export type IntegrationWorkflowsGetResponse = string[];

export const integrationWorkflowsGetQuery = (
  req: IntegrationWorkflowsGetRequest
) => ({
  url: `integration/${req.integrationId}/workflows`,
  headers: { 'api-key': req.apiKey },
});
