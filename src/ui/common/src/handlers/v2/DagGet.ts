// This file should map exactly to
// src/golang/cmd/server/handler/v2/dag_get.go

import { APIKeyParameter } from '../parameters/Header';
import { DagIdParameter, WorkflowIdParameter } from '../parameters/Path';
import { DagResponse } from '../responses/workflow';

export type DagGetRequest = APIKeyParameter &
  DagIdParameter &
  WorkflowIdParameter;

export type DagGetResponse = DagResponse;

export const dagGetQuery = (req: DagGetRequest) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}`,
  headers: { 'api-key': req.apiKey },
});
