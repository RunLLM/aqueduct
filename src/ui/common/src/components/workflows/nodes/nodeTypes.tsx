import {CheckOperatorNode} from "./CheckOperatorNode";
import {MetricOperatorNode} from "./MetricOperatorNode";
import {DatabaseNode} from "./DatabaseNode";
import {FunctionOperatorNode} from "./FunctionOperatorNode";
import {TableArtifactNode} from "./TableArtifactNode";
import {BoolArtifactNode} from "./BoolArtifactNode";
import {FloatArtifactNode} from "./FloatArtifactNode";

export const nodeTypes = {
    database: DatabaseNode,
    tableArtifact: TableArtifactNode,
    floatArtifact: FloatArtifactNode,
    boolArtifact: BoolArtifactNode,
    function: FunctionOperatorNode,

    // These are generic DAG nodes
    functionOp: FunctionOperatorNode,
    extractOp: DatabaseNode,
    loadOp: DatabaseNode,
    metricOp: MetricOperatorNode,
    checkOp: CheckOperatorNode,
};

export default nodeTypes;