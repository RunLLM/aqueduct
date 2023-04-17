// This file should map exactly to
// src/golang/cmd/server/handler/v2/dag_results_get.go

import { APIKeyParameter } from '../parameters/Header';
import { WorkflowIdParameter } from '../parameters/Path';
import { DagResultResponse } from '../responses/Workflow';

export type DagResultsGetRequest = APIKeyParameter & WorkflowIdParameter;

export type DagResultsGetResponse = DagResultResponse[];

export const dagResultsGetQuery = (req: DagResultsGetRequest) => ({
  url: `workflow/${req.workflowId}/results`,
  headers: { 'api-key': req.apiKey },
});
