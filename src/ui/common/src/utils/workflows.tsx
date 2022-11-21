import { Artifact } from './artifacts';
import { EngineConfig } from './engine';
import { normalizeOperator, Operator } from './operators';
import ExecutionStatus, { ExecState } from './shared';
import { StorageConfig } from './storage';

export type S3Config = {
  region: string;
  bucket: string;
};

export enum WorkflowUpdateTrigger {
  Manual = 'manual',
  Periodic = 'periodic',
  Airflow = 'airflow',
}

export type WorkflowSchedule = {
  trigger: WorkflowUpdateTrigger;
  cron_schedule: string;
  disable_manual_trigger: boolean;
  paused: boolean;
};

export type RetentionPolicy = {
  k_latest_runs: number;
};

export type ListWorkflowSummary = {
  id: string;
  name: string;
  description: string;
  created_at: number;
  last_run_at: number;
  status: ExecutionStatus;
  engine: string;
  watcher_auth0_id: string[];
};

export type WorkflowDagResultSummary = {
  id: string;
  created_at: number;
  status: ExecutionStatus;
  workflow_dag_id: string;
};

export type Workflow = {
  id: string;
  user_id: string;
  name: string;
  description: string;
  schedule: WorkflowSchedule;
  created_at: number;
  retention_policy?: RetentionPolicy;
};

export type WorkflowDag = {
  id: string;
  workflow_id: string;
  created_at: number;
  s3_config: S3Config;
  engine_config: EngineConfig;
  storage_config: StorageConfig;

  metadata?: Workflow;
  operators: { [id: string]: Operator };
  artifacts: { [id: string]: Artifact };
};

// This function `normalize` an arbitrary object (typically from an API call)
// to the `WorkflowType` object that actually follows its type definition.
//
// For now, we only handle all lists / maps field. Ideally, we should
// handle all fields like `workflow.id = workflow?.id ?? ''`.
export function normalizeWorkflowDag(dag: WorkflowDag): WorkflowDag {
  const operators: Operator[] = Object.values(dag.operators ?? {});
  dag.operators = {};
  operators.forEach((op) => {
    if (op.id) {
      dag.operators[op.id] = normalizeOperator(op);
    }
  });

  dag.artifacts = dag.artifacts ?? {};
  return dag;
}

export type GetWorkflowResponse = {
  workflow_dags: { [id: string]: WorkflowDag };
  workflow_dag_results: WorkflowDagResultSummary[];
  watcherAuthIds: string[];
};

export type SavedObject = {
  operator_name: string;
  modified_at: string;
  integration_name: string;
  integration_id: string;
  service: string;
  object_name: string;
  update_mode: string;
};

export type ListWorkflowSavedObjectsResponse = {
  object_details: SavedObject[];
};

export type SavedObjectDeletion = {
  name: string;
  exec_state: ExecState;
};

export type DeleteWorkflowResponse = {
  saved_object_deletion_results: { [id: string]: SavedObjectDeletion[] };
};

export function normalizeGetWorkflowResponse(
  resp: GetWorkflowResponse
): GetWorkflowResponse {
  const dags: WorkflowDag[] = Object.values(resp.workflow_dags ?? {});
  resp.workflow_dags = {};
  dags.forEach((dag) => {
    if (dag.id) {
      resp.workflow_dags[dag.id] = normalizeWorkflowDag(dag);
    }
  });
  resp.workflow_dag_results = (resp.workflow_dag_results ?? []).sort((x, y) =>
    x.created_at < y.created_at ? 1 : -1
  );

  return resp;
}

export type ListWorkflowResponse = {
  workflows: ListWorkflowSummary[];
};
