import 'reactflow/dist/style.css';

import React from 'react';
import { useSelector } from 'react-redux';
import ReactFlow, { Node as ReactFlowNode } from 'reactflow';

import { RootState } from '../../stores/store';
import { EdgeTypes, ReactFlowNodeData } from '../../utils/reactflow';
import { ReactflowNodeType } from '../../utils/reactflow';
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

  const selectedDag = useSelector(
    (state: RootState) => state.workflowReducer.selectedDag
  );

  const { edges, nodes } = dagPositionState.result ?? { edges: [], nodes: [] };
  if (edges.length === 0 || nodes.length === 0) {
    // The DAG position state is still loading.
    return null;
  }

  // This is a bit of a tricky check; when we switch between workflow versions, the selected DAG
  // does not load in sync with the DAG positioning. As a result, we need to ensure
  // that we only render the graph once the two sets of Redux state are in sync before
  // proceeding; otherwise, our node IDs will be mismatched. Here, we simply check to see
  // if the UUIDs for one of the nodes exists in the selected DAG. If it doesn't, that
  // means the state has not synced yet, so we return null and wait for it to sync.
  const testNode = nodes[0];
  if (
    (testNode.data.nodeType === ReactflowNodeType.Operator &&
      !selectedDag.operators[testNode.id]) ||
    (testNode.data.nodeType === ReactflowNodeType.Artifact &&
      !selectedDag.artifacts[testNode.id])
  ) {
    return null;
  }

  const defaultViewport = { x: 0, y: 0, zoom: 1 };

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
    // If this is an operator node (which includes metrics & checks),
    // then we should show by default where the operator is running, so we pull
    // that information out of the spec and pass it along.
    const data = { ...node.data };
    if (node.data.nodeType === ReactflowNodeType.Operator) {
      // If an engine config exists on the operator, then that's what we use,
      // but if none exists, we use whatever is the default on the DAG spec.
      data.spec = selectedDag.operators[node.id]?.spec;
      data.dagEngineConfig = selectedDag.engine_config;
    } else {
      data.artifactType = selectedDag.artifacts[node.id]?.type;
    }

    return {
      id: node.id,
      type: node.type,
      data: data,
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
