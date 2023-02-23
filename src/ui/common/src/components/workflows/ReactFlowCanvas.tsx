import React from 'react';
import ReactFlow, { Node as ReactFlowNode } from 'reactflow';
import 'reactflow/dist/style.css';
import { useSelector } from 'react-redux';

import { RootState } from '../../stores/store';
import { EdgeTypes, ReactFlowNodeData } from '../../utils/reactflow';
import nodeTypes from './nodes/nodeTypes';

const connectionLineStyle = { stroke: '#fff' };
const snapGrid = [20, 20];

type ReactFlowCanvasProps = {
  onPaneClicked: (event: React.MouseEvent) => void;
  switchSideSheet: (
    event: React.MouseEvent,
    element: ReactFlowNode<ReactFlowNodeData>
  ) => void;
};

const ReactFlowCanvas: React.FC<ReactFlowCanvasProps> = ({
  onPaneClicked,
  switchSideSheet,
}) => {
  const dagPositionState = useSelector(
    (state: RootState) => state.workflowReducer.selectedDagPosition
  );

  const { edges, nodes } = dagPositionState.result ?? { edges: [], nodes: [] };

  console.log("edges", edges);
  console.log("problem edges", edges[0], edges[1], edges[3], edges[6]);

  const defaultViewport = { x: 0, y: 0, zoom: 1 };
  return (
    <ReactFlow
      onPaneClick={onPaneClicked}
      nodes={nodes}
      edges={[edges[2], edges[4], edges[5]]} // Error: 0, 1, 3, 6
      onNodeClick={switchSideSheet}
      nodeTypes={nodeTypes}
      connectionLineStyle={connectionLineStyle}
      snapToGrid={true}
      snapGrid={snapGrid as [number, number]}
      defaultViewport={defaultViewport}
      edgeTypes={EdgeTypes}
      minZoom={0.25}
    />
  );
};

export default ReactFlowCanvas;
