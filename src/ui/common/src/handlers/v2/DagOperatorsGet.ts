// This file should map exactly to
// src/golang/cmd/server/handler/v2/dag_operators_get.go

import { APIKeyParameter } from '../parameters/Header';
import { DagIdParameter, WorkflowIdParameter } from '../parameters/Path';
import { OperatorResponse } from '../responses/node';

export type DagOperatorsGetRequest = APIKeyParameter &
  DagIdParameter &
  WorkflowIdParameter;

export type DagOperatorsGetResponse = OperatorResponse[];

export const dagOperatorsGetQuery = (req: DagOperatorsGetRequest) => ({
  url: `workflows/${req.workflowId}/dag/${req.dagId}/nodes/operators`,
  headers: { 'api-key': req.apiKey },
});
