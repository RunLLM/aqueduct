// This file should map exactly to
// src/golang/cmd/server/handler/v2/node_check_results_get.go

import { APIKeyParameter } from '../parameters/Header';
import {
  DagIdParameter,
  NodeIdParameter,
  WorkflowIdParameter,
} from '../parameters/Path';
import { OperatorWithArtifactNodeResultResponse } from '../responses/node';

export type NodeCheckResultsGetRequest = APIKeyParameter &
  DagIdParameter &
  NodeIdParameter &
  WorkflowIdParameter;

export type NodeCheckResultsGetResponse =
  OperatorWithArtifactNodeResultResponse[];

export const nodeCheckResultsGetQuery = (req: NodeCheckResultsGetRequest) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}/node/check/${req.nodeId}/results`,
  headers: { 'api-key': req.apiKey },
});
