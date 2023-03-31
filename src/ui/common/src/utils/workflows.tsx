import { Artifact } from './artifacts';
import { EngineConfig, EngineType } from './engine';
import { NotificationLogLevel } from './notifications';
import {
  Load,
  normalizeOperator,
  Operator,
  RelationalDBLoadParams,
  RelationalDBServices,
  S3LoadParams,
  ServiceType,
} from './operators';
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
  Cascade = 'cascade',
}

export type WorkflowSchedule = {
  trigger: WorkflowUpdateTrigger;
  cron_schedule: string;
  disable_manual_trigger: boolean;
  paused: boolean;
  source_id: string;
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
  engine: EngineType;
  operator_engines: EngineType[];
  metrics: WorkflowMetrics[];
  checks: WorkflowChecks[];
};

export type WorkflowDagResultSummary = {
  id: string;
  created_at: number;

  // TODO(ENG-2665): remove the status field.
  status: ExecutionStatus;
  exec_state?: ExecState;
  workflow_dag_id: string;
};

export type NotificationSettingsMap = { [id: string]: NotificationLogLevel };

export type NotificationSettings = { settings: NotificationSettingsMap };

export type Workflow = {
  id: string;
  user_id: string;
  name: string;
  description: string;
  schedule: WorkflowSchedule;
  created_at: number;
  retention_policy?: RetentionPolicy;
  notification_settings?: NotificationSettings;
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
  const serviceString: string = savedObject.spec.service;
  if (Object.values<string>(RelationalDBServices).includes(serviceString)) {
    return (savedObject.spec.parameters as RelationalDBLoadParams).table;
  } else if (savedObject.spec.service == ServiceType.S3) {
    return (savedObject.spec.parameters as S3LoadParams).filepath;
  } else {
    return '';
  }
}

export type SavedObject = {
  operator_name: string;
  modified_at: string;
  integration_name: string;
  spec: Load;
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

export function getWorkflowEngineTypes(dag: WorkflowDag): EngineType[] {
  const dagType = dag.engine_config.type;
  const allOpTypes = Array.from(
    new Set(
      Object.values(dag.operators)
        .map((op) => op.spec.engine_config?.type)
        .filter((t) => !!t) ?? []
    )
  );

  return reduceEngineTypes(dagType, allOpTypes);
}

export function reduceEngineTypes(
  dagType: EngineType,
  opTypes: EngineType[]
): EngineType[] {
  if (
    dagType === EngineType.Aqueduct &&
    opTypes.length > 0 &&
    opTypes.every((t) => t !== EngineType.Aqueduct)
  ) {
    return opTypes;
  }

  return Array.from(new Set([dagType].concat(opTypes)));
}
