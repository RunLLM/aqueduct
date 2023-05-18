// This file should map exactly to
// src/golang/cmd/server/handler/v2/resources_workflows_get.go

import { APIKeyParameter } from '../parameters/Header';
import { WorkflowAndDagIDResponse } from '../responses/workflow';

export type IntegrationsWorkflowsGetRequest = APIKeyParameter;

// IntegrationID -> list of workflows that use this integration.
export type IntegrationsWorkflowsGetResponse = {
  [integrationID: string]: WorkflowAndDagIDResponse[];
};

export const integrationsWorkflowsGetQuery = (
  req: IntegrationsWorkflowsGetRequest
) => ({
  url: `resources/workflows`,
  headers: { 'api-key': req.apiKey },
});
