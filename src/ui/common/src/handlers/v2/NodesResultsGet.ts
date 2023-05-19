// This file should map exactly to
// src/golang/cmd/server/handler/v2/nodes_results_get.go

import { APIKeyParameter } from '../parameters/Header';
import { DagResultIdParameter, WorkflowIdParameter } from '../parameters/Path';
import { NodeResultsResponse } from '../responses/node';

export type NodesResultsGetRequest = APIKeyParameter &
  DagResultIdParameter &
  WorkflowIdParameter;

export type NodesResultsGetResponse = NodeResultsResponse;

export const nodesResultsGetQuery = (req: NodesResultsGetRequest) => ({
  url: `workflow/${req.workflowId}/result/${req.dagResultId}/nodes/results`,
  headers: { 'api-key': req.apiKey },
});
