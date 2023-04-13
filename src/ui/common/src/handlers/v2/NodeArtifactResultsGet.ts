// This file should map exactly to
// src/golang/cmd/server/handler/v2/node_artifact_results_get.go

import { APIKeyParameter } from '../parameters/ApiKey';
import { DagIdParameter } from '../parameters/DagId';
import { NodeIdParameter } from '../parameters/NodeId';
import { WorkflowIdParameter } from '../parameters/WorkflowId';
import { ArtifactResultResponse } from '../responses/Node';

export type NodeArtifactResultsGetRequest = APIKeyParameter &
  DagIdParameter &
  NodeIdParameter &
  WorkflowIdParameter;

export type NodeArtifactResultsGetResponse = ArtifactResultResponse[];

export const nodeArtifactResultsGetQuery = (
  req: NodeArtifactResultsGetRequest
) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}/node/artifact/${req.nodeId}/results`,
  headers: { 'api-key': req.apiKey },
});
