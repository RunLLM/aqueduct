import Box from '@mui/material/Box';
import React from 'react';
import { Node } from 'reactflow';

import ArtifactDetailsPage from '../components/pages/artifact/id';
import CheckDetailsPage from '../components/pages/check/id';
import MetricDetailsPage from '../components/pages/metric/id';
import OperatorDetailsPage from '../components/pages/operator/id';
import { ArtifactResponse, OperatorResponse } from '../handlers/responses/node';
import { ReactFlowNodeData } from '../positioning/positioning';
import { NodeType, selectNode } from '../reducers/nodeSelection';
import { NodeSelection } from '../reducers/pages/Workflow';
import { AppDispatch } from '../stores/store';
import UserProfile from './auth';
import { OperatorType } from './operators';

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
  selectedNodeState: NodeSelection,
  selectedNode: OperatorResponse | ArtifactResponse
): React.ReactElement {
  const SideSheetLayout = ({ children }) => {
    return (
      <Box px={'16px'} maxWidth="800px" height="100vh">
        {children}
      </Box>
    );
  };

  if (selectedNodeState.nodeType === 'artifacts') {
    return (
      <ArtifactDetailsPage
        user={user}
        Layout={SideSheetLayout}
        nodeId={selectedNodeState.nodeId}
        sideSheetMode={true}
      />
    );
  }

  const opNode = selectedNode as OperatorResponse;
  if (opNode.spec?.type === OperatorType.Metric) {
    return (
      <MetricDetailsPage
        user={user}
        Layout={SideSheetLayout}
        nodeId={selectedNodeState.nodeId}
        sideSheetMode={true}
      />
    );
  }

  if (opNode.spec?.type === OperatorType.Check) {
    return (
      <CheckDetailsPage
        user={user}
        Layout={SideSheetLayout}
        nodeId={selectedNodeState.nodeId}
        sideSheetMode={true}
      />
    );
  }

  return (
    <OperatorDetailsPage
      user={user}
      Layout={SideSheetLayout}
      nodeId={selectedNodeState.nodeId}
      sideSheetMode={true}
    />
  );
}
