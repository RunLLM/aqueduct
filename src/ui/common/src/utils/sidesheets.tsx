import Box from '@mui/material/Box';
import React from 'react';
import { Node } from 'reactflow';

import ArtifactDetailsPage from '../components/pages/artifact/id';
import CheckDetailsPage from '../components/pages/check/id';
import MetricDetailsPage from '../components/pages/metric/id';
import OperatorDetailsPage from '../components/pages/operator/id';
import { NodeType, SelectedNode, selectNode } from '../reducers/nodeSelection';
import { AppDispatch } from '../stores/store';
import UserProfile from './auth';
import { ReactFlowNodeData } from './reactflow';

/**
 * This function takes in a dispatch call (which must be created in a
 * component) and a call to a set state function in the using component, and it
 * returns a function which takes an event for a click on a node in ReactFlow
 * and opens the appropriate corresponding sidesheet.
 */
export const sideSheetSwitcher = (dispatch: AppDispatch) => {
  return (event: React.MouseEvent, element: Node<ReactFlowNodeData>): void => {
    dispatch(selectNode({ id: element.id, type: element.type as NodeType }));
  };
};

export function getDataSideSheetContent(
  user: UserProfile,
  currentNode: SelectedNode,
  workflowIdProp: string,
  workflowDagIdProp: string | undefined,
  workflowDagResultIdProp: string | undefined
): React.ReactElement {
  const SideSheetLayout = ({ children }) => {
    return (
      <Box px={'16px'} maxWidth="800px" height="100vh">
        {children}
      </Box>
    );
  };

  switch (currentNode.type) {
    case NodeType.BoolArtifact:
    case NodeType.NumericArtifact:
    case NodeType.TableArtifact:
    case NodeType.JsonArtifact:
    case NodeType.StringArtifact:
    case NodeType.ImageArtifact:
    case NodeType.DictArtifact:
    case NodeType.ListArtifact:
    case NodeType.GenericArtifact:
      return (
        <ArtifactDetailsPage
          user={user}
          Layout={SideSheetLayout}
          artifactIdProp={currentNode.id}
          workflowDagIdProp={workflowDagIdProp}
          workflowDagResultIdProp={workflowDagResultIdProp}
          workflowIdProp={workflowIdProp}
          sideSheetMode={true}
        />
      );
    case NodeType.CheckOp:
      return (
        <CheckDetailsPage
          user={user}
          Layout={SideSheetLayout}
          operatorIdProp={currentNode.id}
          workflowDagIdProp={workflowDagIdProp}
          workflowDagResultIdProp={workflowDagResultIdProp}
          workflowIdProp={workflowIdProp}
          sideSheetMode={true}
        />
      );
    case NodeType.MetricOp:
      return (
        <MetricDetailsPage
          user={user}
          Layout={SideSheetLayout}
          operatorIdProp={currentNode.id}
          workflowDagIdProp={workflowDagIdProp}
          workflowDagResultIdProp={workflowDagResultIdProp}
          workflowIdProp={workflowIdProp}
          sideSheetMode={true}
        />
      );
    case NodeType.ExtractOp:
    case NodeType.LoadOp:
    case NodeType.FunctionOp: {
      return (
        <OperatorDetailsPage
          user={user}
          Layout={SideSheetLayout}
          operatorIdProp={currentNode.id}
          workflowDagIdProp={workflowDagIdProp}
          workflowDagResultIdProp={workflowDagResultIdProp}
          workflowIdProp={workflowIdProp}
          sideSheetMode={true}
        />
      );
    }
  }
}
