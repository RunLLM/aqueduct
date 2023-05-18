import { Divider } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';
import { useParams } from 'react-router-dom';

import DefaultLayout from '../../../../components/layouts/default';
import LogViewer from '../../../../components/LogViewer';
import WithOperatorHeader from '../../../../components/operators/WithOperatorHeader';
import UserProfile from '../../../../utils/auth';
import OperatorSpecDetails from '../../../workflows/operator/specDetails';
import { LayoutProps } from '../../types';
import {
  useWorkflowBreadcrumbs,
  useWorkflowIds,
  useWorkflowNodes,
  useWorkflowNodesResults,
} from '../../workflow/id/hook';

type OperatorDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
  nodeId?: string;
  sideSheetMode?: boolean;
};

// Checked with file size=313285391 and handles that smoothly once loaded. However, takes a while to load.
const OperatorDetailsPage: React.FC<OperatorDetailsPageProps> = ({
  user,
  Layout = DefaultLayout,
  nodeId,
  sideSheetMode = false,
}) => {
  const { workflowId, dagId, dagResultId } = useWorkflowIds(user.apiKey);

  const { nodeId: nodeIdParam } = useParams();
  if (!nodeId) {
    nodeId = nodeIdParam;
  }

  const breadcrumbs = useWorkflowBreadcrumbs(
    user.apiKey,
    workflowId,
    dagId,
    dagResultId,
    'Operator'
  );

  const nodes = useWorkflowNodes(user.apiKey, workflowId, dagId);
  const nodeResults = useWorkflowNodesResults(
    user.apiKey,
    workflowId,
    dagResultId
  );

  const node = nodes.operators[nodeId];
  const nodeResult = nodeResults.operators[nodeId];

  const logs = nodeResult?.exec_state?.user_logs ?? {};
  const operatorError = nodeResult?.exec_state?.error;

  return (
    <Layout breadcrumbs={breadcrumbs} user={user}>
      <WithOperatorHeader
        workflowId={workflowId}
        dagId={dagId}
        dagResultId={dagResultId}
        nodes={nodes}
        nodeResults={nodeResults}
        operator={node}
        operatorResult={nodeResult}
        sideSheetMode={sideSheetMode}
      >
        <Box>
          <Typography variant="h6" fontWeight="normal">
            Logs
          </Typography>
          {!!logs && <LogViewer logs={logs} err={operatorError} />}
        </Box>

        <Divider sx={{ my: '32px' }} />
        <OperatorSpecDetails user={user} operator={node} />
      </WithOperatorHeader>
    </Layout>
  );
};

export default OperatorDetailsPage;
