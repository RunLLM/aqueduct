// This file should map exactly to
// src/golang/cmd/server/handler/v2/resource_operators_get.go

import { APIKeyParameter } from '../parameters/Header';
import { ResourceIdParameter } from '../parameters/Path';
import { OperatorResponse } from '../responses/node';

export type ResourceOperatorsGetRequest = APIKeyParameter &
  ResourceIdParameter;

export type ResourceOperatorsGetResponse = OperatorResponse[];

export const resourceOperatorsGetQuery = (
  req: ResourceOperatorsGetRequest
) => ({
  url: `resource/${req.resourceId}/nodes/operators`,
  headers: { 'api-key': req.apiKey },
});
