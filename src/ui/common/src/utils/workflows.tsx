import { Artifact } from './artifacts';
import { EngineConfig } from './engine';
import {Load, normalizeOperator, Operator, RelationalDBLoadParams, S3LoadParams} from './operators';
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

export type WorkflowMetrics = {
  id: string;
  description: string;
  from: string;
  name: string;
  to: string;
  result: {
    content_path: string;
    // This is the thing that we want to show in the table view.
    content_serialized: string;
    exec_state: ExecState;
    serialization_type: string;
  };
};

export type ExecutionResult = {
  id: string;
  exec_state: ExecState;
};

export type WorkflowChecks = {
  id: string;
  description: string;
  // inputs: need to figure out what goes in there.
  // outputs: need to figure this out too.
  name: string;
  result: ExecutionResult;
  spec: {
    check: {
      level: string;
      function: {
        custom_args: string;
        granularity: string;
        language: string;
        storage_path: string;
        type: string;
      };
    };
    type: string;
  };
};

export type ListWorkflowSummary = {
  id: string;
  name: string;
  description: string;
  created_at: number;
  last_run_at: number;
  status: ExecutionStatus;
  engine: string;
  metrics: WorkflowMetrics[];
  checks: WorkflowChecks[];
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

  // The default engine that this workflow was run with. Can be overriden by individual operator specs.
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
};

export function getSavedObjectIdentifier(savedObject: SavedObject): string {
  if (savedObject.spec.parameters as S3LoadParams !== undefined) {
    return (savedObject.spec.parameters as S3LoadParams).filepath
  } else if (savedObject.spec.parameters as RelationalDBLoadParams !== undefined) {
    return (savedObject.spec.parameters as RelationalDBLoadParams).table
  } else {
    return "";
  }
}

export type SavedObject = {
  operator_name: string;
  modified_at: string;
  integration_name: string;
  spec: Load;

  // TODO: REMOVE
  object_name: string; // Used in delete_workflow
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
