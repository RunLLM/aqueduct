// This file should map exactly to
// src/golang/cmd/server/handler/v2/nodes_results_get.go

import { APIKeyParameter } from '../parameters/ApiKey';
import { DagResultIdParameter } from '../parameters/DagResultId';
import { WorkflowIdParameter } from '../parameters/WorkflowId';
import { NodeResultsResponse } from '../responses/Node';

export type NodesResultsGetRequest = APIKeyParameter &
  DagResultIdParameter &
  WorkflowIdParameter;

export type NodesResultsGetResponse = NodeResultsResponse;

export const nodesResultsGetQuery = (req: NodesResultsGetRequest) => ({
  url: `workflow/${req.workflowId}/result/${req.dagResultId}/nodes/results`,
  headers: { 'api-key': req.apiKey },
});
