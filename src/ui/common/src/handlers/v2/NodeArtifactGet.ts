// This file should map exactly to
// src/golang/cmd/server/handler/v2/node_artifact_get.go

import { APIKeyParameter } from '../parameters/Header';
import {
  DagIdParameter,
  NodeIdParameter,
  WorkflowIdParameter,
} from '../parameters/Path';
import { ArtifactResponse } from '../responses/Node';

export type NodeArtifactGetRequest = APIKeyParameter &
  DagIdParameter &
  NodeIdParameter &
  WorkflowIdParameter;

export type NodeArtifactGetResponse = ArtifactResponse;

export const nodeArtifactGetQuery = (req: NodeArtifactGetRequest) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}/node/artifact/${req.nodeId}`,
  headers: { 'api-key': req.apiKey },
});
