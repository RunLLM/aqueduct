import { Artifact } from './artifacts';
import { EngineConfig } from './engine';
import { normalizeOperator, Operator } from './operators';
import ExecutionStatus, { ExecState } from './shared';

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

  metadata?: Workflow;
  operators: { [id: string]: Operator };
  artifacts: { [id: string]: Artifact };
};

// This function `normalize` an arbitrary object (typically from an API call)
// to the `WorkflowType` object that actually follows its type definition.
//
// For now, we only handle all lists / maps field. Ideally, we should
// handle all fields like `workflow.id = workflow?.id ?? ''`.
export function normalizeWorkflowDag(dag): WorkflowDag {
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

export function normalizeGetWorkflowResponse(resp): GetWorkflowResponse {
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

/**
 * This is a simplified version of Sugiyama method https://blog.disy.net/sugiyama-method/
 * where we compute the DAG topology for rendering its layout.
 *
 * It takes a list of operators and returns the following:
 * - Layers, which is a list of layers, where each layer is a list of opId
 * - Number of active edges for each layer. 'active edges' stands for any edges at this layer, that doesn't starts
 *   from the previous layer, nor ends at current layer. This is used to compute how much `room` one layer
 *   need to save for rendering edges.
 *
 */
export function computeTopologicalOrder(operators: {
  [id: string]: Operator;
}): [string[][], number[]] {
  const artifactToDownstream: { [id: string]: string[] } = {};
  const upstreamCount: { [id: string]: number } = {};
  const layers: string[][] = [];
  const activeLayerEdges: number[] = [];
  let activeEdges = 0;
  layers.push([]);
  activeLayerEdges.push(0);

  for (const opId in operators) {
    const op = operators[opId];
    op.inputs.map((artfId) => {
      if (!(artfId in artifactToDownstream)) {
        artifactToDownstream[artfId] = [];
      }
      artifactToDownstream[artfId].push(opId);
    });

    upstreamCount[opId] = op.inputs.length;
    if (op.inputs.length === 0) {
      layers[layers.length - 1].push(opId);
    }
  }

  while (layers[layers.length - 1].length > 0) {
    const frontier = layers[layers.length - 1];
    layers.push([]);
    frontier.map((opId) => {
      const op = operators[opId];
      op.outputs.map((artfId) => {
        if (artfId in artifactToDownstream) {
          artifactToDownstream[artfId].map((downstreamOpId) => {
            activeEdges += 1;
            upstreamCount[downstreamOpId] = upstreamCount[downstreamOpId] - 1;
            if (upstreamCount[downstreamOpId] === 0) {
              layers[layers.length - 1].push(downstreamOpId);
              activeEdges -= operators[downstreamOpId].inputs.length;
            }
          });
        }
      });
    });
    activeLayerEdges.push(activeEdges);
  }
  return [layers, activeLayerEdges];
}
