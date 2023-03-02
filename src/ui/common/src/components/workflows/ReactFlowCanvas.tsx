import React from 'react';
import ReactFlow, {
  Controls,
  MiniMap,
  Node as ReactFlowNode,
} from 'react-flow-renderer';
import { useSelector } from 'react-redux';

import { RootState } from '../../stores/store';
import { EdgeTypes, ReactFlowNodeData } from '../../utils/reactflow';
import nodeTypes from './nodes/nodeTypes';

const connectionLineStyle = { stroke: '#fff' };

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

  const canvasEdges = edges.map((edge) => {
    return {
      id: edge.id,
      source: edge.source,
      target: edge.target,
      type: edge.type,
      container: 'root',
    };
  });

  const canvasNodes = nodes.map((node) => {
    return {
      id: node.id,
      type: node.type,
      data: node.data,
      position: node.position,
    };
  });

  return (
    <ReactFlow
      onPaneClick={onPaneClicked}
      nodes={canvasNodes}
      edges={canvasEdges}
      onNodeClick={switchSideSheet}
      nodeTypes={nodeTypes}
      connectionLineStyle={connectionLineStyle}
      edgeTypes={EdgeTypes}
      onInit={(reactFlowInstance) => {
        console.log('onInitCalled, ', reactFlowInstance);
        //reactFlowInstance.fitBounds({ x: 0, y: 0, width: 0, height: 0 });
        //reactFlowInstance.fitView();
        reactFlowInstance.setViewport({ x: 50, y: 50, zoom: 0.4 });
      }}
      onMove={(event, viewport) => {
        console.log('onMove viewport: ', viewport);
      }}
      defaultZoom={2}
      minZoom={0.2}
    >
      <MiniMap />
      <Controls />
    </ReactFlow>
  );
};

export default ReactFlowCanvas;
