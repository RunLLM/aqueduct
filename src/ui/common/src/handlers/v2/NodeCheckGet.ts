// This file should map exactly to
// src/golang/cmd/server/handler/v2/node_check_get.go

import { OperatorWithArtifactNodeResponse } from '../responses/Node';
import { NodeArtifactGetRequest } from './NodeArtifactGet';

export type NodeCheckGetRequest = NodeArtifactGetRequest;

export type NodeCheckGetResponse = OperatorWithArtifactNodeResponse;

export const nodeCheckGetQuery = (req: NodeCheckGetRequest) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}/node/check/${req.nodeId}`,
  headers: { 'api-key': req.apiKey },
});
