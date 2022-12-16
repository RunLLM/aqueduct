import { Edge, Node } from 'react-flow-renderer';
import { Position } from 'react-flow-renderer';

import AqueductBezier from '../components/workflows/edges/AqueductBezier';
import AqueductQuadratic from '../components/workflows/edges/AqueductQuadratic';
import AqueductStraight from '../components/workflows/edges/AqueductStraight';
import {
  ArtifactTypeToNodeTypeMap,
  OperatorTypeToNodeTypeMap,
} from '../reducers/nodeSelection';
import { Artifact } from './artifacts';
import { Operator } from './operators';

/**
 * Type representing the various lines used to connect DAG nodes to one another on the DAG view.
 */
export const EdgeTypes = {
  quadratic: AqueductQuadratic,
  straight: AqueductStraight,
  curved: AqueductBezier,
};

/**
 * Type representing the various nodes shown on the DAG view.
 */
export enum ReactflowNodeType {
  Operator = 'operator',
  Artifact = 'artifact',
}

/**
 * Type representing the data contained in each DAG node on the DAG view.
 */
export type ReactFlowNodeData = {
  /**
   * Callback function to call when the node changes.
   * @returns void
   */
  onChange: () => void;
  /**
   * Callback function to be called when the user drags a line from one node's handle to another.
   * NOTE: Currently we do not support editing the DAG, so this function is not used.
   * @param any input parameters needed (if any) to be called when two DAG nodes are connected to one another
   * @returns void.
   */
  onConnect: (any) => void;
  /**
   * Where to place the handles used to position connecting edge lines.
   * Handles are also used to allow users to drag and drop a line connecting two nodes.
   * This functionality is currently unsupported in Aqueduct.
   */
  targetHandlePosition?: Position;
  /**
   * The type of the node to be rendered on the DAG.
   */
  nodeType: ReactflowNodeType;
  /**
   * Unique identifier of the node.
   */
  nodeId: string;
  /**
   * Label text to show on the node.
   */
  label?: string;
  /**
   * Used to present metric or check results inside the node.
   */
  result?: string;
};

/**
 * Response type used to store operator and artifact x and y positions on the DAG viewer.
 */
export type GetPositionResponse = {
  operator_positions: { [opId: string]: NodePos };
  artifact_positions: { [artfId: string]: NodePos };
};

/**
 * x and y position for a node on the DAG viewer.
 */
type NodePos = { x: number; y: number };

/**
 * Converts an Aqueduct operator node into a react-flow node.
 * @param op The operator to be rendered.
 * @param pos x and y position of the operator.
 * @param onChange Callback function to be called when the operator node changes.
 * @param onConnect Callback function to be called when user drags and drops one node's handle to another. This functionality is currently not used.
 * @returns A react flow node representing an Aqueduct Operator to be rendered in the DAG view.
 */
export function getOperatorNode(
  op: Operator,
  pos: NodePos,
  onChange: () => void,
  onConnect: (any) => void
): Node<ReactFlowNodeData> {
  return {
    id: op.id,
    type: OperatorTypeToNodeTypeMap[op.spec.type],
    draggable: false,
    data: {
      nodeType: ReactflowNodeType.Operator,
      onChange,
      onConnect,
      nodeId: op.id,
      label: op.name,
    },
    position: pos,
  };
}

/**
 * Converts an Aqueduct artifact node into a react-flow node.
 * @param artf The artifact to be rendered.
 * @param pos x and y position of the artifact.
 * @param onChange Callback function to be called when the artifact node changes.
 * @param onConnect Callback function to be called when user drags and drops one node's handle to another. This functionality is currently not used.
 * @returns A react-flow ndoe representing an Aqueduct Artifact to be rendered in the DAG view.
 */
export function getArtifactNode(
  artf: Artifact,
  pos: NodePos,
  onChange: () => void,
  onConnect: (any) => void
): Node<ReactFlowNodeData> {
  return {
    id: artf.id,
    type: ArtifactTypeToNodeTypeMap[artf.type],
    draggable: false,
    data: {
      nodeType: ReactflowNodeType.Artifact,
      onChange,
      onConnect,
      nodeId: artf.id,
      label: artf.name,
    },
    position: pos,
  };
}

/**
 * Retrieves a list of edges for a given set of operators.
 * @param operators Map of operators to retrieve DAG edges for,
 * @returns An array of input and output edges for each operator of the operators map.
 */
export function getEdges(operators: { [id: string]: Operator }): Edge[] {
  const results = [];
  Object.values(operators).forEach((op) => {
    op.inputs.forEach((artfId) => {
      results.push({
        id: `${artfId}-${op.id}`,
        type: 'curved', // Op inputs are curved edges
        source: artfId,
        target: op.id,
      });
    });
    op.outputs.forEach((artfId) => {
      results.push({
        id: `${op.id}-${artfId}`,
        type: 'straight',
        source: op.id,
        target: artfId,
      });
    });
  });
  return results;
}
