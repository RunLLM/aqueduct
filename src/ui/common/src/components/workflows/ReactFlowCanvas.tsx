import { produce } from 'immer';
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

  const artifactResults = useSelector(
    (state: RootState) => state.workflowReducer.artifactResults
  );

  const currentNode = useSelector(
    (state: RootState) => state.nodeSelectionReducer.selected
  );

  const { fitView } = useReactFlow();
  useEffect(() => {
    setTimeout(fitView, 1000);
  }, [dagPositionState]);

  useEffect(() => {
    // NOTE(vikram): There's a timeout here because there seems to be a
    // race condition between calling `fitView` and the viewport
    // updating. This might be because of the width transition we use, but
    // we're not 100% sure.
    setTimeout(fitView, 100);
  }, [currentNode]);

  const collapseNodes = () => {
    const checkOpNodes = [];
    const boolArtifactNodes = [];

    const metricOpNodes = [];
    const metricArtifactNodes = [];

    // first find all check operators.
    if (dagPositionState.result) {
      const nodes = dagPositionState.result.nodes;
      nodes.forEach((node) => {
        if (node.type === 'checkOp') {
          checkOpNodes.push(node);
        } else if (node.type === 'boolArtifact') {
          boolArtifactNodes.push(node);
        } else if (node.type === 'metricOp') {
          metricOpNodes.push(node);
        } else if (node.type === 'numericArtifact') {
          metricArtifactNodes.push(node);
        }
      });

      //sort checkOpNodes and boolArtifactNodes by value of label
      //operator nodes have just a name
      //artifact nodes have same name + ' artifact' at the end.
      const alphabeticallySortNodes = (a, b) =>
        a.data.label.localeCompare(b.data.label);
      checkOpNodes.sort(alphabeticallySortNodes);
      boolArtifactNodes.sort(alphabeticallySortNodes);

      // Do the same sorting for metric and metric artifact nodes.
      metricOpNodes.sort(alphabeticallySortNodes);
      metricArtifactNodes.sort(alphabeticallySortNodes);
    }

    // Remove artifactNodes from the DAG
    // Take artifact result and set inside operator's data.result field.
    const nodes = dagPositionState.result?.nodes;
    let enrichedNodes = [];

    if (nodes) {
      enrichedNodes = produce(nodes, (draftState) => {
        // NOTE: only mutate the draftState variable here.
        // See docs here for more information: https://redux-toolkit.js.org/usage/immer-reducers#immutable-updates-with-immer
        if (nodes) {
          // loop through and find checkOpNodes, doing fancy logic stuffs.
          for (let nodeIndex = 0; nodeIndex < nodes.length; nodeIndex++) {
            // boolArtifactNodes and checkOps are sorted and now have the same index as one another.
            // Let's take the operators and set their data.result fields accordingly.
            for (let i = 0; i < checkOpNodes.length; i++) {
              if (nodes[nodeIndex].id === checkOpNodes[i].id) {
                // Let's find the artifact result of the corresponding booleanArtifactNode.
                const boolArtifactResult =
                  artifactResults[boolArtifactNodes[i].id]?.result?.data;
                if (boolArtifactResult) {
                  draftState[nodeIndex].data.result = boolArtifactResult;
                }
              }
            }

            // find metric nodes and put the result into the operator's data field
            for (let i = 0; i < metricOpNodes.length; i++) {
              if (nodes[nodeIndex].id === metricOpNodes[i].id) {
                const metricArtifactResult =
                  artifactResults[metricArtifactNodes[i].id]?.result?.data;
                if (metricArtifactResult) {
                  draftState[nodeIndex].data.result = metricArtifactResult;
                }
              }
            }
          }
        }
      });
    }

    //Finally, let's remove any boolean artifacts from the list
    // This has to be two separate steps or Immer will complain that we are producing a new value and modifying it's draft.
    // i.e. Error: [Immer] An immer producer returned a new value *and* modified its draft. Either return a new value *or* modify the draft.
    const filteredNodes = produce(enrichedNodes, (draftState) => {
      if (!enrichedNodes) {
        return [];
      }

      return draftState.filter((node) => {
        for (let i = 0; i < boolArtifactNodes.length; i++) {
          if (node.id === boolArtifactNodes[i].id) {
            return false;
          }
        }

        for (let i = 0; i < metricArtifactNodes.length; i++) {
          if (node.id === metricArtifactNodes[i].id) {
            return false;
          }
        }

        return true;
      });
    });

    const edges = dagPositionState.result?.edges;
    const updatedEdges = produce(edges, (edgeDraftState) => {
      if (!edges) {
        return [];
      }

      // we have two sorted arrays of metric and metric artifact nodes.
      // each array entry corresponds to an entry at the same index in the other array.
      // metricOpNode                         metricArtifactNode
      for (let i = 0; i < metricOpNodes.length; i++) {
        const metricOpNode = metricOpNodes[i];
        const metricArtifactNode = metricArtifactNodes[i];

        // find all edges with the metricArtifactNode as a source.
        for (let edgeIndex = 0; edgeIndex < edges.length; edgeIndex++) {
          if (edges[edgeIndex].source === metricArtifactNode.id) {
            edgeDraftState[edgeIndex].source = metricOpNode.id;
          }
        }
      }
    });

    // remove any edge who's target is a metric artifact.
    const filteredEdges = produce(updatedEdges, (draftState) => {
      if (!updatedEdges) {
        return [];
      }

      return draftState.filter((edge) => {
        for (let i = 0; i < metricArtifactNodes.length; i++) {
          if (
            edge.target === metricArtifactNodes[i].id ||
            edge.source === metricArtifactNodes[i].id
          ) {
            return false;
          }
        }

        return true;
      });
    });

    return {
      edges: filteredEdges,
      nodes: filteredNodes,
    };
  };

  const { edges, nodes } = collapseNodes();

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
