// This file should map exactly to
// `src/golang/lib/response/workflow.go`

import { EngineConfig } from '../../utils/engine';
import { ExecState } from '../../utils/shared';
import { StorageConfig } from '../../utils/storage';
import {
  NotificationSettings,
  RetentionPolicy,
  WorkflowSchedule,
} from '../../utils/workflows';

export type WorkflowResponse = {
  id: string;
  user_id: string;
  name: string;
  description: string;
  schedule: WorkflowSchedule;
  created_at: string;
  retention_policy: RetentionPolicy;
  notification_settings: NotificationSettings;
};

export type DagResponse = {
  id: string;
  workflow_id: string;
  created_at: string;
  storage_config: StorageConfig;
  engine_config: EngineConfig;
};

export type DagResultResponse = {
  id: string;
  dag_id: string;
  exec_state: ExecState;
};

export type WorkflowAndDagIDResponse = {
  id: string;
  dag_id: string;
};
