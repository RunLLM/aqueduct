// This file should map exactly to
// src/golang/cmd/server/handler/v2/dag_get.go

import { APIKeyParameter } from '../parameters/ApiKey';
import { DagIdParameter } from '../parameters/DagId';
import { WorkflowIdParameter } from '../parameters/WorkflowId';
import { DagResponse } from '../responses/Workflow';

export type DagGetRequest = APIKeyParameter &
  DagIdParameter &
  WorkflowIdParameter;

export type DagGetResponse = DagResponse;

export const dagGetQuery = (req: DagGetRequest) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}`,
  headers: { 'api-key': req.apiKey },
});
