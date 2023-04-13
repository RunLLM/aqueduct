// This file should map exactly to
// src/golang/cmd/server/handler/v2/node_artifact_result_content_get.go

import { APIKeyParameter } from '../parameters/ApiKey';
import { DagIdParameter } from '../parameters/DagId';
import { NodeIdParameter } from '../parameters/NodeId';
import { NodeResultIdParameter } from '../parameters/NodeResultId';
import { WorkflowIdParameter } from '../parameters/WorkflowId';

export type NodeArtifactResultContentGetRequest = APIKeyParameter &
  DagIdParameter &
  NodeIdParameter &
  WorkflowIdParameter &
  NodeResultIdParameter;

export type NodeArtifactResultContentGetResponse = {
  is_downsampled: boolean;
  content: string;
};

export const nodeArtifactResultContentGetQuery = (
  req: NodeArtifactResultContentGetRequest
) => ({
  url: `workflow/${req.workflowId}/dag/${req.dagId}/node/artifact/${req.nodeId}/result/${req.nodeResultId}/content`,
  headers: { 'api-key': req.apiKey },
});
