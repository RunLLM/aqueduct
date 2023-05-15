// This file should map exactly to
// src/golang/cmd/server/handler/v2/node_check_result_content_get.go

import { NodeArtifactResultContentGetRequest, NodeArtifactResultContentGetResponse } from './NodeArtifactResultContentGet';

export type NodeCheckResultContentGetRequest = NodeArtifactResultContentGetRequest;

export type NodeCheckResultContentGetResponse = NodeArtifactResultContentGetResponse;

export const nodeCheckResultContentGetQuery = (
  req: NodeCheckResultContentGetRequest
) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}/node/check/${req.nodeId}/result/${req.nodeResultId}/content`,
  headers: { 'api-key': req.apiKey },
});
