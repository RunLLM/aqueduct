// This file should map exactly to
// src/golang/cmd/server/handler/v2/integration_workflows.go

import { APIKeyParameter } from '../parameters/Header';
import { IntegrationIdParameter } from '../parameters/Path';

export type IntegrationOperatorsGetRequest = APIKeyParameter &
  IntegrationIdParameter;

// A list of workflow IDs that use this integration.
export type IntegrationOperatorsGetResponse = string[];

export const integrationOperatorsGetQuery = (
  req: IntegrationOperatorsGetRequest
) => ({
  url: `integration/${req.integrationId}/nodes/operators`,
  headers: { 'api-key': req.apiKey },
});
