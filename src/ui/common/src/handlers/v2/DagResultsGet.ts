// This file should map exactly to
// src/golang/cmd/server/handler/v2/dag_results_get.go

import { APIKeyParameter } from '../parameters/Header';
import { WorkflowIdParameter } from '../parameters/Path';
import {
  LimitParameter,
  OrderByParameter,
  OrderDescParameter,
} from '../parameters/Query';
import { DagResultResponse } from '../responses/workflow';

export type DagResultsGetRequest = APIKeyParameter &
  WorkflowIdParameter &
  LimitParameter &
  OrderByParameter &
  OrderDescParameter;

export type DagResultsGetResponse = DagResultResponse[];

export const dagResultsGetQuery = (req: DagResultsGetRequest) => ({
  url: `workflow/${req.workflowId}/results`,
  headers: { 'api-key': req.apiKey },
});
