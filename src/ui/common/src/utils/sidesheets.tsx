import Box from '@mui/material/Box';
import React from 'react';

import ArtifactDetailsPage from '../components/pages/artifact/id';
import CheckDetailsPage from '../components/pages/check/id';
import MetricDetailsPage from '../components/pages/metric/id';
import OperatorDetailsPage from '../components/pages/operator/id';
import { ArtifactResponse, OperatorResponse } from '../handlers/responses/node';
import { NodeSelection } from '../reducers/pages/Workflow';
import UserProfile from './auth';
import { OperatorType } from './operators';

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
