// This file should mirror src/golang/workflow/dag/response.go
import { EngineConfig } from '../../utils/engine';
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
