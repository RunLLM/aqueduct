import { Edge, Node } from 'reactflow';
import { Position } from 'reactflow';

import AqueductBezier from '../components/workflows/edges/AqueductBezier';
import AqueductQuadratic from '../components/workflows/edges/AqueductQuadratic';
import AqueductStraight from '../components/workflows/edges/AqueductStraight';
import {
  ArtifactTypeToNodeTypeMap,
  OperatorTypeToNodeTypeMap,
} from '../reducers/nodeSelection';
import { Artifact } from './artifacts';
import { Operator } from './operators';

export const EdgeTypes = {
  quadratic: AqueductQuadratic,
  straight: AqueductStraight,
  curved: AqueductBezier,
};

export enum ReactflowNodeType {
  Operator = 'operator',
  Artifact = 'artifact',
}

export type ReactFlowNodeData = {
  onChange: () => void;
  onConnect: (any) => void;
  targetHandlePosition?: Position;
  nodeType: ReactflowNodeType;
  nodeId: string;
  label?: string;
  // Used to present metric or check results inside the node
  result?: string;
};

export type GetPositionResponse = {
  operator_positions: { [opId: string]: NodePos };
  artifact_positions: { [artfId: string]: NodePos };
};

// These are generic dag supports
type NodePos = { x: number; y: number };

export function getOperatorNode(
  op: Operator,
  pos: NodePos,
  onChange: () => void,
  onConnect: (any) => void,
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
    // Give an initial position. We will reposition this node later.
    //position: { x: 0, y: 0 },
    position: pos,
  };
}

export function getArtifactNode(
  artf: Artifact,
  pos: NodePos,
  onChange: () => void,
  onConnect: (any) => void,
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
    // Give an initial position. We will reposition this node later.
    //position: { x: 0, y: 0 },
    position: pos
  };
}

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
