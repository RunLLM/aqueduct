import { ArtifactType } from '../../../utils/artifacts';
import { OperatorType } from '../../../utils/operators';
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
import MetricOperatorNode from './MetricOperatorNode';
import { metricOperatorNodeIcon } from './MetricOperatorNode';
import NumericArtifactNode, {
  numericArtifactNodeIcon,
} from './NumericArtifactNode';
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
  genericArtifact: GenericArtifactNode,
  function: FunctionOperatorNode,

  // These are generic DAG nodes
  functionOp: FunctionOperatorNode,
  extractOp: DatabaseNode,
  loadOp: DatabaseNode,
  metricOp: MetricOperatorNode,
  checkOp: CheckOperatorNode,
};

export const artifactTypeToIconMapping = {
  [ArtifactType.String]: stringArtifactNodeIcon,
  [ArtifactType.Bool]: boolArtifactNodeIcon,
  [ArtifactType.Numeric]: numericArtifactNodeIcon,
  [ArtifactType.Dict]: dictArtifactNodeIcon,
  // TODO: figure out if we should use other icon for tuple
  [ArtifactType.Tuple]: dictArtifactNodeIcon,
  [ArtifactType.Table]: tableArtifactNodeIcon,
  [ArtifactType.Json]: jsonArtifactNodeIcon,
  // TODO: figure out what to show for bytes.
  [ArtifactType.Bytes]: dictArtifactNodeIcon,
  [ArtifactType.Image]: imageArtifactNodeIcon,
  // TODO: Figure out what to show for Picklable
  [ArtifactType.Picklable]: dictArtifactNodeIcon,
};

export const operatorTypeToIconMapping = {
  [OperatorType.Function]: functionOperatorNodeIcon,
  [OperatorType.Extract]: databaseNodeIcon,
  [OperatorType.Load]: databaseNodeIcon,
  [OperatorType.Metric]: metricOperatorNodeIcon,
  [OperatorType.Check]: checkOperatorNodeIcon,
  [OperatorType.SystemMetric]: metricOperatorNodeIcon,
};

export default nodeTypes;
