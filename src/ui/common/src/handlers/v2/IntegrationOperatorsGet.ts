// This file should map exactly to
// src/golang/cmd/server/handler/v2/integration_operators_get.go

import { APIKeyParameter } from '../parameters/Header';
import { IntegrationIdParameter } from '../parameters/Path';
import { OperatorResponse } from '../responses/node';

export type IntegrationOperatorsGetRequest = APIKeyParameter &
  IntegrationIdParameter;

export type IntegrationOperatorsGetResponse = OperatorResponse[];

export const integrationOperatorsGetQuery = (
  req: IntegrationOperatorsGetRequest
) => ({
  url: `integration/${req.integrationId}/nodes/operators`,
  headers: { 'api-key': req.apiKey },
});
