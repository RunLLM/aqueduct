// This file should map exactly to
// src/golang/cmd/server/handler/v2/node_metric_get.go

import { MergedNodeResponse } from '../responses/Node';
import { NodeArtifactGetRequest } from './NodeArtifactGet';

export type NodeMetricGetRequest = NodeArtifactGetRequest;

export type NodeMetricGetResponse = MergedNodeResponse;

export const nodeMetricGetQuery = (req: NodeMetricGetRequest) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}/node/metric/${req.nodeId}`,
  headers: { 'api-key': req.apiKey },
});
