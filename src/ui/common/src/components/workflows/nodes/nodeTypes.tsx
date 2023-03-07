import { ArtifactType } from '../../../utils/artifacts';
import { OperatorType } from '../../../utils/operators';
import ExecutionStatus from '../../../utils/shared';
import BoolArtifactNode, { boolArtifactNodeIcon } from './BoolArtifactNode';
import CheckOperatorNode from './CheckOperatorNode';
import { checkOperatorNodeIcon } from './CheckOperatorNode';
import DatabaseNode from './DatabaseNode';
import { databaseNodeIcon } from './DatabaseNode';
import DictArtifactNode, { dictArtifactNodeIcon } from './DictArtifactNode';
import FunctionOperatorNode from './FunctionOperatorNode';
import { functionOperatorNodeIcon } from './FunctionOperatorNode';
import GenericArtifactNode from './GenericArtifactNode';
import ImageArtifactNode, { imageArtifactNodeIcon } from './ImageArtifactNode';
import JsonArtifactNode, { jsonArtifactNodeIcon } from './JsonArtifactNode';
import ListArtifactNode, { listArtifactNodeIcon } from './ListArtifactNode';
import MetricOperatorNode from './MetricOperatorNode';
import { metricOperatorNodeIcon } from './MetricOperatorNode';
import NumericArtifactNode, {
  numericArtifactNodeIcon,
} from './NumericArtifactNode';
import ParameterOperatorNode, {
  paramOperatorNodeIcon,
} from './ParameterOperatorNode';
import StringArtifactNode, {
  stringArtifactNodeIcon,
} from './StringArtifactNode';
import TableArtifactNode, { tableArtifactNodeIcon } from './TableArtifactNode';

export const nodeTypes = {
  database: DatabaseNode,
  tableArtifact: TableArtifactNode,
  numericArtifact: NumericArtifactNode,
  boolArtifact: BoolArtifactNode,
  jsonArtifact: JsonArtifactNode,
  stringArtifact: StringArtifactNode,
  imageArtifact: ImageArtifactNode,
  dictArtifact: DictArtifactNode,
  listArtifact: ListArtifactNode,
  genericArtifact: GenericArtifactNode,
  function: FunctionOperatorNode,

  // These are generic DAG nodes
  functionOp: FunctionOperatorNode,
  extractOp: DatabaseNode,
  loadOp: DatabaseNode,
  metricOp: MetricOperatorNode,
  checkOp: CheckOperatorNode,
  paramOp: ParameterOperatorNode,
};

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
  [ArtifactType.String]: stringArtifactNodeIcon,
  [ArtifactType.Bool]: boolArtifactNodeIcon,
  [ArtifactType.Numeric]: numericArtifactNodeIcon,
  [ArtifactType.Dict]: dictArtifactNodeIcon,
  // TODO: figure out if we should use other icon for tuple
  [ArtifactType.Tuple]: dictArtifactNodeIcon,
  [ArtifactType.List]: listArtifactNodeIcon,
  [ArtifactType.Table]: tableArtifactNodeIcon,
  [ArtifactType.Json]: jsonArtifactNodeIcon,
  // TODO: figure out what to show for bytes.
  [ArtifactType.Bytes]: dictArtifactNodeIcon,
  [ArtifactType.Image]: imageArtifactNodeIcon,
  // TODO: Figure out what to show for Picklable
  [ArtifactType.Picklable]: dictArtifactNodeIcon,
};

export const operatorTypeToIconMapping = {
  [OperatorType.Param]: paramOperatorNodeIcon,
  [OperatorType.Function]: functionOperatorNodeIcon,
  [OperatorType.Extract]: databaseNodeIcon,
  [OperatorType.Load]: databaseNodeIcon,
  [OperatorType.Metric]: metricOperatorNodeIcon,
  [OperatorType.Check]: checkOperatorNodeIcon,
  [OperatorType.SystemMetric]: metricOperatorNodeIcon,
};

export default nodeTypes;
