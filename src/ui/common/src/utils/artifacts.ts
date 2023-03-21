import { WorkflowState } from '../reducers/workflow';
import ExecutionStatus, { ExecState } from './shared';

export enum ArtifactType {
  String = 'string',
  Bool = 'boolean',
  Numeric = 'numeric',
  Dict = 'dictionary',
  Tuple = 'tuple',
  List = 'list',
  Table = 'table',
  Json = 'json',
  Bytes = 'bytes',
  Image = 'image',
  Picklable = 'picklable',
  Untyped = 'untyped',
}

export enum SerializationType {
  String = 'string',
  BsonTable = 'bson_table',
  Table = 'table',
  Json = 'json',
  Bytes = 'bytes',
  Image = 'image',
  Pickle = 'pickle',
}

export type Artifact = {
  id: string;
  name: string;
  description: string;
  type: ArtifactType;
};

export type Schema = { [col_name: string]: string }[];

export type ArtifactResultContent = {
  data?: string;
  is_downsampled: boolean;
};

export type GetArtifactResultResponse = ArtifactResultContent & {
  name: string;
  // `status` is technically redundant due to `execState`. Avoid using `status` in new code.
  status: ExecutionStatus;
  exec_state: ExecState;
  schema: Schema;
  artifact_type: ArtifactType;
  serialization_type: SerializationType;
  // TODO: python_type goes here.
};

// Takes the ID of an artifact in our DAG and the state of the currently
// selected workflow and returns the ID of the operator that is responsible for
// creating the artifact.
export const getUpstreamOperator = (
  workflowState: WorkflowState,
  artifactId: string
): string => {
  let result: string;

  // Load all operators in the current dag version and iterate through
  // them.
  const operators = workflowState.selectedDag?.operators;

  for (const operator of Object.values(operators)) {
    // Check if this particular operator was the one responsible for
    // creating the artifact we care about. If so, pull the
    // `GetOperatorResultResponse` from the `workflowReducer` state and
    // retrieve the error from that response.
    if (operator.outputs.includes(artifactId)) {
      result = operator.id;
      break;
    }
  }

  return result;
};
