import 'reactflow/dist/style.css';

import React from 'react';
import { useDispatch } from 'react-redux';
import ReactFlow, { Node as ReactFlowNode } from 'reactflow';

import {
  NodeResultsMap,
  NodesMap,
  NodesResponse,
} from '../../handlers/responses/node';
import { DagResponse } from '../../handlers/responses/workflow';
import { ReactFlowNodeData, visualizeDag } from '../../positioning/positioning';
import { selectNode } from '../../reducers/pages/Workflow';
import AqueductBezier from './edges/AqueductBezier';
import AqueductQuadratic from './edges/AqueductQuadratic';
import AqueductStraight from './edges/AqueductStraight';
import Node from './nodes/Node';

const connectionLineStyle = { stroke: '#fff' };
const snapGrid = [20, 20];

const NodeTypes = {
  operators: Node,
  artifacts: Node,
};

const EdgeTypes = {
  quadratic: AqueductQuadratic,
  straight: AqueductStraight,
  curved: AqueductBezier,
};

type ReactFlowCanvasProps = {
  nodes: NodesMap;
  nodeResults?: NodeResultsMap;
  dag: DagResponse;
};

const ReactFlowCanvas: React.FC<ReactFlowCanvasProps> = ({
  nodes,
  nodeResults,
  dag,
}) => {
  const dispatch = useDispatch();
  const visualizedDag = visualizeDag(dag, nodes, nodeResults);

  const defaultViewport = { x: 0, y: 0, zoom: 1 };

  return (
    <ReactFlow
      onPaneClick={(event: React.MouseEvent) => {
        event.preventDefault();

        // Reset selected node
        dispatch(
          selectNode({ workflowId: dag.workflow_id, selection: undefined })
        );
      }}
      nodes={visualizedDag.nodes}
      edges={visualizedDag.edges}
      onNodeClick={(
        event: React.MouseEvent,
        element: ReactFlowNode<ReactFlowNodeData>
      ) => {
        dispatch(
          selectNode({
            workflowId: dag.workflow_id,
            selection: {
              nodeId: element.id,
              nodeType: element.type as keyof NodesResponse,
            },
          })
        );
      }}
      nodeTypes={NodeTypes}
      connectionLineStyle={connectionLineStyle}
      snapToGrid={true}
      snapGrid={snapGrid as [number, number]}
      defaultViewport={defaultViewport}
      edgeTypes={EdgeTypes}
      minZoom={0.25}
      fitView={true}
    />
  );
};

export default ReactFlowCanvas;
