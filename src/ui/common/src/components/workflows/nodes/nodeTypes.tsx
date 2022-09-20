import BoolArtifactNode from './BoolArtifactNode';
import CheckOperatorNode from './CheckOperatorNode';
import DatabaseNode from './DatabaseNode';
import DictArtifactNode from './DictArtifactNode';
import FunctionOperatorNode from './FunctionOperatorNode';
import GenericArtifactNode from './GenericArtifactNode';
import ImageArtifactNode from './ImageArtifactNode';
import JsonArtifactNode from './JsonArtifactNode';
import MetricOperatorNode from './MetricOperatorNode';
import NumericArtifactNode from './NumericArtifactNode';
import StringArtifactNode from './StringArtifactNode';
import TableArtifactNode from './TableArtifactNode';

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

export default nodeTypes;
