// This file should mirror src/golang/workflow/dag/response.go
import { EngineConfig } from '../../utils/engine';
import { ExecutionStatus } from '../../utils/shared';
import { StorageConfig } from '../../utils/storage';
import {
  WorkflowRetentionPolicy,
  WorkflowSchedule,
} from '../../utils/workflows';
import { ArtifactResponse, ArtifactResultResponse } from './artifact';
import { OperatorResponse, OperatorResultResponse } from './operator';

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
  retention_policy?: WorkflowRetentionPolicy;
};

export type DagResponse = DagMetadataResponse & {
  operators: { [id: string]: OperatorResponse };
  artifacts: { [id: string]: ArtifactResponse };
};

export type DagRawResultResponse = {
  id: string;
  status: ExecutionStatus;
  created_at: number;
};

export type DagResultResponse = DagMetadataResponse & {
  result?: DagRawResultResponse;
  operators: { [id: string]: OperatorResultResponse };
  artifacts: { [id: string]: ArtifactResultResponse };
};
