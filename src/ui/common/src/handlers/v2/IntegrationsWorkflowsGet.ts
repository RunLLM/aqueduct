// This file should map exactly to
// src/golang/cmd/server/handler/v2/integrations_workflows.go

import { APIKeyParameter } from '../parameters/Header';

export type IntegrationsWorkflowsGetRequest = APIKeyParameter;

// IntegrationID -> list of workflows that use this integration.
export type IntegrationsWorkflowsGetResponse = {
  [integrationID: string]: string[];
};

export const integrationsWorkflowsGetQuery = (
  req: IntegrationsWorkflowsGetRequest
) => ({
  url: `integrations/workflows`,
  headers: { 'api-key': req.apiKey },
});
