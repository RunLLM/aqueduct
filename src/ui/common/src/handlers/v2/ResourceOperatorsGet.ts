// This file should map exactly to
// src/golang/cmd/server/handler/v2/resource_operators_get.go

import { APIKeyParameter } from '../parameters/Header';
import { IntegrationIdParameter } from '../parameters/Path';
import { OperatorResponse } from '../responses/node';

export type IntegrationOperatorsGetRequest = APIKeyParameter &
  IntegrationIdParameter;

export type IntegrationOperatorsGetResponse = OperatorResponse[];

export const resourceOperatorsGetQuery = (
  req: IntegrationOperatorsGetRequest
) => ({
  url: `resource/${req.resourceId}/nodes/operators`,
  headers: { 'api-key': req.apiKey },
});
