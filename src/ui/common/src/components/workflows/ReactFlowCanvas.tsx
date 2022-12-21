import React, { useEffect } from 'react';
import ReactFlow, {
  Node as ReactFlowNode,
  useReactFlow,
} from 'react-flow-renderer';
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

  const { fitView } = useReactFlow();

  useEffect(() => {
    fitView();
  }, [dagPositionState, fitView]);

  const { edges, nodes } = dagPositionState.result ?? { edges: [], nodes: [] };

  return (
    <ReactFlow
      onPaneClick={onPaneClicked}
      nodes={nodes}
      edges={edges}
      onNodeClick={switchSideSheet}
      nodeTypes={nodeTypes}
      connectionLineStyle={connectionLineStyle}
      snapToGrid={true}
      snapGrid={snapGrid as [number, number]}
      defaultZoom={1}
      edgeTypes={EdgeTypes}
      minZoom={0.25}
    />
  );
};

export default ReactFlowCanvas;
