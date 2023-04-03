// This file should mirror src/golang/workflow/dag/response.go
import { EngineConfig } from '../../utils/engine';
import { OperatorType } from '../../utils/operators';
import { ExecState } from '../../utils/shared';
import { StorageConfig } from '../../utils/storage';
import { RetentionPolicy, WorkflowSchedule } from '../../utils/workflows';
import { ArtifactResponse, ArtifactResultResponse } from './artifactDeprecated';
import { OperatorResponse, OperatorResultResponse } from './operatorDeprecated';

export type DagMetadataResponse = {
  dag_id: string;
  dag_created_at: number;
  storage_config?: StorageConfig;
  engine_config?: EngineConfig;

  workflow_id: string;
  workflow_created_at: number;
  user_id: string;
  name: string;
  description: string;
  schedule?: WorkflowSchedule;
  retention_policy?: RetentionPolicy;
};

export type DagResponse = DagMetadataResponse & {
  operators: { [id: string]: OperatorResponse };
  artifacts: { [id: string]: ArtifactResponse };
};

export type DagResultStatusResponse = {
  id: string;
  exec_state?: ExecState;
};

export type DagResultResponse = DagMetadataResponse & {
  result?: DagResultStatusResponse;
  operators: { [id: string]: OperatorResultResponse };
  artifacts: { [id: string]: ArtifactResultResponse };
};

// This helper fetches all metrics and checks defined on an artifact. Which includes:
// - metrics with this artifact as input
// - checks with this artifact, and all above metrics' outputs as input
export function getMetricsAndChecksOnArtifact(
  dagResult: DagResultResponse,
  artifactId: string
): { checks: ArtifactResultResponse[]; metrics: ArtifactResultResponse[] } {
  const metricsOp = Object.values(dagResult.operators).filter(
    (opResult) =>
      opResult.inputs.includes(artifactId) &&
      (opResult.spec?.type === OperatorType.Metric ||
        opResult.spec?.type === OperatorType.SystemMetric)
  );
  const checksOp = Object.values(dagResult.operators).filter(
    (opResult) =>
      opResult.inputs.includes(artifactId) &&
      opResult.spec?.type === OperatorType.Check
  );

  const metricsArtfIds = metricsOp.flatMap((opResult) => {
    return opResult !== undefined ? opResult.outputs : [];
  });

  const metricsArtf = metricsArtfIds.map((id) => dagResult.artifacts[id]);
  const metricsDownstreamIds = metricsArtf.flatMap(
    (artfResult) => artfResult.to
  );

  const metricsDownstreamOps = metricsDownstreamIds.map(
    (id) => dagResult.operators[id]
  );

  checksOp.push(...metricsDownstreamOps);

  const checksArtfIds = checksOp.flatMap((opResult) =>
    opResult !== undefined ? opResult.outputs : []
  );

  const checksArtf = checksArtfIds.map((id) => dagResult.artifacts[id]);

  return { checks: checksArtf, metrics: metricsArtf };
}
