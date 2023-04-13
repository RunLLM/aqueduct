// This file should map exactly to
// src/golang/cmd/server/handler/v2/dag_result_get.go

import { APIKeyParameter } from '../parameters/ApiKey';
import { DagResultIdParameter } from '../parameters/DagResultId';
import { WorkflowIdParameter } from '../parameters/WorkflowId';
import { DagResultResponse } from '../responses/Workflow';

export type DagResultGetRequest = APIKeyParameter &
  DagResultIdParameter &
  WorkflowIdParameter;

export type DagResultGetResponse = DagResultResponse;

export const dagResultGetQuery = (req: DagResultGetRequest) => ({
  url: `workflow/${req.workflowId}/result/${req.dagResultId}`,
  headers: { 'api-key': req.apiKey },
});
