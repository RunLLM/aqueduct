// This file should map exactly to
// `src/golang/lib/response/node.go`
import { ArtifactType, SerializationType } from '../../utils/artifacts';
import { CheckLevel, OperatorSpec, OperatorType } from '../../utils/operators';
import ExecutionStatus, { ExecState } from '../../utils/shared';

export type OperatorWithArtifactNodeResponse = {
  id: string;
  dag_id: string;
  artifact_id?: string;
  name: string;
  description: string;
  spec?: OperatorSpec;
  type?: ArtifactType;
  inputs: string[];
  outputs: string[];
};

export type OperatorWithArtifactNodeResultResponse = {
  id: string;
  operator_exec_state?: ExecState;

  artifact_id?: string;
  serialization_type?: SerializationType;
  content_path?: string;
  content_serialized?: string;
  artifact_exec_state?: ExecState;
};

export type ArtifactResponse = {
  id: string;
  dag_id: string;
  name: string;
  description: string;
  type: ArtifactType;
  input: string;
  outputs: string[];
};

export type ArtifactResultResponse = {
  id: string;
  artifact_id: string;
  serialization_type: SerializationType;
  content_path: string;
  content_serialized: string;
  exec_state?: ExecState;
};

export type OperatorResponse = {
  id: string;
  dag_id: string;
  name: string;
  description: string;
  spec?: OperatorSpec;
  inputs: string[];
  outputs: string[];
};

export type OperatorResultResponse = {
  id: string;
  operator_id: string;
  exec_state?: ExecState;
};

export type NodesResponse = {
  operators: OperatorResponse[];
  artifacts: ArtifactResponse[];
  // TODO: ENG-2987 Create separate sections for Metrics/Checks
  // metrics: OperatorWithArtifactNodeResponse[];
  // checks: OperatorWithArtifactNodeResponse[];
};

export type NodeResultsResponse = {
  operators: OperatorResultResponse[];
  artifacts: ArtifactResultResponse[];
  // TODO: ENG-2987 Create separate sections for Metrics/Checks
  // metrics: OperatorWithArtifactNodeResultResponse[];
  // checks: OperatorWithArtifactNodeResultResponse[];
};

export type NodesMap = {
  operators: { [id: string]: OperatorResponse };
  artifacts: { [id: string]: ArtifactResponse };
};

export type NodeResultsMap = {
  operators: { [id: string]: OperatorResultResponse };
  artifacts: { [id: string]: ArtifactResultResponse };
};

export type NodeContentResponse = {
  name: string;
  data: string;
};

export function hasWarningCheck(
  workflowStatus: ExecutionStatus,
  nodes: NodesMap,
  nodesResults: NodeResultsMap
): boolean {
  if (workflowStatus && workflowStatus === ExecutionStatus.Succeeded) {
    const ops = Object.values(nodes.operators);
    for (let i = 0; i < ops.length; i++) {
      const op = ops[i];
      if (
        op.spec.type === 'check' &&
        op.spec.check.level === CheckLevel.Warning
      ) {
        const artifactId = op.outputs[0]; // Assuming there is only one output artifact
        const value = nodesResults.artifacts[artifactId]?.exec_state?.status;
        if (value === ExecutionStatus.Failed) {
          return true;
        }
      }
    }
  }
  return false;
}

export function getMetricsAndChecksOnArtifact(
  nodes: NodesMap,
  artifactId: string
): { checks: OperatorResponse[]; metrics: OperatorResponse[] } {
  const metricsOp = Object.values(nodes.operators).filter(
    (op) =>
      op.inputs.includes(artifactId) &&
      (op.spec?.type === OperatorType.Metric ||
        op.spec?.type === OperatorType.SystemMetric)
  );
  const checksOp = Object.values(nodes.operators).filter(
    (op) =>
      op.inputs.includes(artifactId) && op.spec?.type === OperatorType.Check
  );

  const metricsArtfIds = metricsOp.flatMap((op) => {
    return op !== undefined ? op.outputs : [];
  });

  const metricsArtf = metricsArtfIds.map((id) => nodes.artifacts[id]);
  const metricsDownstreamIds = metricsArtf.flatMap((artf) => artf.outputs);

  const metricsDownstreamOps = metricsDownstreamIds.map(
    (id) => nodes.operators[id]
  );

  checksOp.push(...metricsDownstreamOps);

  return { checks: checksOp, metrics: metricsOp };
}
