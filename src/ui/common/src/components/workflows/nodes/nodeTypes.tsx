import {
  faCheck,
  faCircleCheck,
  faCode,
  faDatabase,
  faFileCode,
  faFileText,
  faHashtag,
  faImage,
  faList,
  faPencil,
  faSliders,
  faTableColumns,
  faTemperatureHalf,
} from '@fortawesome/free-solid-svg-icons';

import { ArtifactType } from '../../../utils/artifacts';
import { OperatorType } from '../../../utils/operators';
import ExecutionStatus from '../../../utils/shared';
import Node from './Node';

export const nodeTypes = {
  database: Node,
  tableArtifact: Node,
  numericArtifact: Node,
  boolArtifact: Node,
  jsonArtifact: Node,
  stringArtifact: Node,
  imageArtifact: Node,
  dictArtifact: Node,
  listArtifact: Node,
  genericArtifact: Node,
  function: Node,

  // These are generic DAG nodes
  functionOp: Node,
  extractOp: Node,
  loadOp: Node,
  metricOp: Node,
  checkOp: Node,
  paramOp: Node,
};

// TODO: Double check that this can be deprecated.
export const nodeTypeToStringLabel = {
  tableArtifact: 'Table Artifact',
  numericArtifact: 'Numeric Artifact',
  boolArtifact: 'Boolean Artifact',
  jsonArtifact: 'JSON Artifact',
  stringArtifact: 'String Artifact',
  imageArtifact: 'Image Artifact',
  dictArtifact: 'Dictionary Artifact',
  listArtifact: 'List Artifact',
  genericArtifact: 'Generic Artifact',
  // NOTE function and functionOp are the same. Should remove one in the future?
  function: 'Function Operator',
  functionOp: 'Function Operator',
  extractOp: 'Extract Operator',
  loadOp: 'Load Operator',
  metricOp: 'Metric Operator',
  checkOp: 'Check Operator',
  paramOp: 'Parameter',
};

export const artifactNodeStatusLabels = {
  [ExecutionStatus.Succeeded]: 'Created',
  [ExecutionStatus.Failed]: 'Failed',
  [ExecutionStatus.Pending]: 'Pending',
  [ExecutionStatus.Canceled]: 'Canceled',
  [ExecutionStatus.Registered]: 'Registered',
  [ExecutionStatus.Running]: 'Running',
  [ExecutionStatus.Warning]: 'Warning',
  [ExecutionStatus.Unknown]: 'Unknown',
};

export const operatorNodeStatusLabels = {
  [ExecutionStatus.Succeeded]: 'Succeeded',
  [ExecutionStatus.Failed]: 'Errored',
  [ExecutionStatus.Pending]: 'Pending',
  [ExecutionStatus.Canceled]: 'Canceled',
  [ExecutionStatus.Registered]: 'Registered',
  [ExecutionStatus.Running]: 'Running',
  [ExecutionStatus.Warning]: 'Warning',
  [ExecutionStatus.Unknown]: 'Unknown',
};

export const metricNodeStatusLabels = operatorNodeStatusLabels;

export const checkNodeStatusLabels = {
  [ExecutionStatus.Succeeded]: 'Passed',
  [ExecutionStatus.Failed]: 'Failed',
  [ExecutionStatus.Pending]: 'Pending',
  [ExecutionStatus.Canceled]: 'Canceled',
  [ExecutionStatus.Registered]: 'Registered',
  [ExecutionStatus.Running]: 'Running',
  [ExecutionStatus.Warning]: 'Warning',
  [ExecutionStatus.Unknown]: 'Unknown',
};
export const artifactTypeToIconMapping = {
  [ArtifactType.String]: faFileText,
  [ArtifactType.Bool]: faCircleCheck,
  [ArtifactType.Numeric]: faHashtag,
  [ArtifactType.Dict]: faFileCode,
  // TODO: figure out if we should use other icon for tuple
  [ArtifactType.Tuple]: faFileCode,
  [ArtifactType.List]: faList,
  [ArtifactType.Table]: faTableColumns,
  [ArtifactType.Json]: faPencil,
  // TODO: figure out what to show for bytes.
  [ArtifactType.Bytes]: faFileCode,
  [ArtifactType.Image]: faImage,
  // TODO: Figure out what to show for Picklable
  [ArtifactType.Picklable]: faFileCode,
  [ArtifactType.Untyped]: faPencil,
};

export const operatorTypeToIconMapping = {
  [OperatorType.Param]: faSliders,
  [OperatorType.Function]: faCode,
  [OperatorType.Extract]: faDatabase,
  [OperatorType.Load]: faDatabase,
  [OperatorType.Metric]: faTemperatureHalf,
  [OperatorType.Check]: faCheck,
  [OperatorType.SystemMetric]: faTemperatureHalf,
};

export default nodeTypes;
