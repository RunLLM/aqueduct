// This file should map exactly to
// src/golang/cmd/server/handler/v2/dag_get.go

import { APIKeyRequest } from './requests/ApiKey';
import { ArtifactResponse } from './responses/node';

export type NodeArtifactGetRequest = APIKeyRequest & {
  workflowId: string;
  dagId: string;
  nodeId: string;
};

export type NodeArtifactGetResponse = ArtifactResponse;

export const nodeArtifactGetQuery = (req: NodeArtifactGetRequest) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}/node/artifact/${req.nodeId}`,
  headers: { 'api-key': req.apiKey },
});
