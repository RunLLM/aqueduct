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
  const openSideSheetState = useSelector(
    (state: RootState) => state.openSideSheetReducer
  );
  const dagPositionState = useSelector(
    (state: RootState) => state.workflowReducer.selectedDagPosition
  );

  const { fitView, viewportInitialized } = useReactFlow();
  useEffect(() => {
    fitView();
  }, [dagPositionState]);

  useEffect(() => {
    // NOTE(vikram): There's a timeout here because there seems to be a
    // race condition between calling `fitView` and the viewport
    // updating. This might be because of the width transition we use, but
    // we're not 100% sure.
    setTimeout(fitView, 200);
  }, [openSideSheetState]);

  //console.log('nodes: ', dagPositionState.result.nodes);
  console.log('dagPositionState: ', dagPositionState);

  const checkOpNodes = [];
  const boolArtifactNodes = [];
  // first find all check operators.
  if (dagPositionState.result) {


    const nodes = dagPositionState.result.nodes;
    nodes.forEach(node => {
      if (node.type === 'checkOp') {
        checkOpNodes.push(node);
      } else if (node.type === 'boolArtifact') {
        boolArtifactNodes.push(node);
      }
    });

    //sort checkOpNodes and boolArtifactNodes by value of label
    //operator nodes have just a name
    //artifact nodes have same name + ' artifact' at the end.
    checkOpNodes.sort((a, b) => a.data.label.localeCompare(b.data.label));
    boolArtifactNodes.sort((a, b) => a.data.label.localeCompare(b.data.label));

    console.log('checkOpNodes', checkOpNodes);
    console.log('boolArtifactNodes: ', boolArtifactNodes);
  }

  // Remove artifactNodes from the DAG
  let nodes = dagPositionState.result?.nodes;
  if (nodes) {
    nodes = nodes.filter((node) => {
      for (let i = 0; i < boolArtifactNodes.length; i++) {
        if (node.id === boolArtifactNodes[i].id) {
          return false
        }
      }

      return true;
    })
  }

  // Remove edges that went into each boolArtifactNode

  return (
    <ReactFlow
      onPaneClick={onPaneClicked}
      nodes={nodes}
      edges={dagPositionState.result?.edges}
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
