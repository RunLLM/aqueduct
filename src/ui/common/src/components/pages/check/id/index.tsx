import Box from '@mui/material/Box';
import React from 'react';
import { useParams } from 'react-router-dom';

import WithOperatorHeader from '../../../../components/operators/WithOperatorHeader';
import { useNodeArtifactResultsGetQuery } from '../../../../handlers/AqueductApi';
import UserProfile from '../../../../utils/auth';
import DefaultLayout from '../../../layouts/default';
import CheckHistory from '../../../workflows/artifact/check/history';
import { LayoutProps } from '../../types';
import {
  useWorkflowBreadcrumbs,
  useWorkflowIds,
  useWorkflowNodes,
  useWorkflowNodesResults,
} from '../../workflow/id/hook';

type CheckDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
  nodeId?: string;
  sideSheetMode?: boolean;
};

const CheckDetailsPage: React.FC<CheckDetailsPageProps> = ({
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
        {!!history && (
          <Box
            width={sideSheetMode ? 'auto' : '49.2%'}
            marginTop={sideSheetMode ? '16px' : '40px'}
          >
            <CheckHistory
              history={history}
              isLoading={isLoading}
              error={error as string}
              checkLevel={node?.spec?.check?.level}
            />
          </Box>
        )}
      </WithOperatorHeader>
    </Layout>
  );
};

export default CheckDetailsPage;
