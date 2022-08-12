import BoolArtifactNode from './BoolArtifactNode';
import CheckOperatorNode from './CheckOperatorNode';
import DatabaseNode from './DatabaseNode';
import NumericArtifactNode from './NumericArtifactNode';
import FunctionOperatorNode from './FunctionOperatorNode';
import JsonArtifactNode from './JsonArtifactNode';
import MetricOperatorNode from './MetricOperatorNode';
import TabularArtifactNode from './TabularArtifactNode';

export const nodeTypes = {
  database: DatabaseNode,
  tabularArtifact: TabularArtifactNode,
  numericArtifact: NumericArtifactNode,
  boolArtifact: BoolArtifactNode,
  jsonArtifact: JsonArtifactNode,
  function: FunctionOperatorNode,

  // These are generic DAG nodes
  functionOp: FunctionOperatorNode,
  extractOp: DatabaseNode,
  loadOp: DatabaseNode,
  metricOp: MetricOperatorNode,
  checkOp: CheckOperatorNode,
};

export default nodeTypes;
