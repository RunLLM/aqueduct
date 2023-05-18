import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';
import { useParams } from 'react-router-dom';

import WithOperatorHeader from '../../../../components/operators/WithOperatorHeader';
import { useNodeArtifactResultsGetQuery } from '../../../../handlers/AqueductApi';
import UserProfile from '../../../../utils/auth';
import DefaultLayout from '../../../layouts/default';
import LogViewer from '../../../LogViewer';
import MetricsHistory from '../../../workflows/artifact/metric/history';
import { LayoutProps } from '../../types';
import {
  useWorkflowBreadcrumbs,
  useWorkflowIds,
  useWorkflowNodes,
  useWorkflowNodesResults,
} from '../../workflow/id/hook';

type MetricDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
  nodeId?: string;
  sideSheetMode?: boolean;
};

const MetricDetailsPage: React.FC<MetricDetailsPageProps> = ({
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

  const artifactId = node?.outputs[0];
  const {
    data: history,
    isLoading,
    error,
  } = useNodeArtifactResultsGetQuery({
    apiKey: user.apiKey,
    nodeId: artifactId,
    workflowId,
    dagId,
  });

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
        {!!logs && (
          <Box>
            <Typography variant="h6" fontWeight="normal">
              Logs
            </Typography>
            <LogViewer logs={logs} err={operatorError} />
          </Box>
        )}
        {!!history && (
          <Box
            width={sideSheetMode ? 'auto' : '49.2%'}
            marginTop={sideSheetMode ? '16px' : '40px'}
          >
            <MetricsHistory
              history={history}
              isLoading={isLoading}
              error={error as string}
            />
          </Box>
        )}
      </WithOperatorHeader>
    </Layout>
  );
};

export default MetricDetailsPage;
