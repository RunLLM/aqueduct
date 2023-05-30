// This file should map exactly to
// src/golang/cmd/server/handler/v2/dags_get.go

import { APIKeyParameter } from '../parameters/Header';
import { WorkflowIdParameter } from '../parameters/Path';
import { DagResponse } from '../responses/workflow';

export type DagsGetRequest = APIKeyParameter & WorkflowIdParameter;

export type DagsGetResponse = DagResponse[];

export const dagsGetQuery = (req: DagsGetRequest) => ({
  url: `workflow/${req.workflowId}/dags`,
  headers: { 'api-key': req.apiKey },
});
