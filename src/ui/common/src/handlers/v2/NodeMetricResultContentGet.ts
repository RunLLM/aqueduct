// This file should map exactly to
// src/golang/cmd/server/handler/v2/node_metric_result_content_get.go

import {
  NodeArtifactResultContentGetRequest,
  NodeArtifactResultContentGetResponse,
} from './NodeArtifactResultContentGet';

export type NodeMetricResultContentGetRequest =
  NodeArtifactResultContentGetRequest;

export type NodeMetricResultContentGetResponse =
  NodeArtifactResultContentGetResponse;

export const nodeMetricResultContentGetQuery = (
  req: NodeMetricResultContentGetRequest
) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}/node/metric/${req.nodeId}/result/${req.nodeResultId}/content`,
  headers: { 'api-key': req.apiKey },
});
