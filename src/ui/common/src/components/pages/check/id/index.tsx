import Box from '@mui/material/Box';
import React from 'react';

import WithOperatorHeader from '../../../../components/operators/WithOperatorHeader';
import UserProfile from '../../../../utils/auth';
import DefaultLayout from '../../../layouts/default';
import RequireOperator from '../../../operators/RequireOperator';
import CheckHistory from '../../../workflows/artifact/check/history';
import RequireDagOrResult from '../../../workflows/RequireDagOrResult';
import { useArtifactHistory } from '../../artifact/id/hook';
import useOpeartor from '../../operator/id/hook';
import { LayoutProps } from '../../types';
import useWorkflow from '../../workflow/id/hook';

type CheckDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
  workflowIdProp?: string;
  workflowDagIdProp?: string;
  workflowDagResultIdProp?: string;
  operatorIdProp?: string;
  sideSheetMode?: boolean;
};

const CheckDetailsPage: React.FC<CheckDetailsPageProps> = ({
  user,
  Layout = DefaultLayout,
  workflowIdProp,
  workflowDagIdProp,
  workflowDagResultIdProp,
  operatorIdProp,
  sideSheetMode = false,
}) => {
  const {
    breadcrumbs: wfBreadcrumbs,
    workflowId,
    workflowDagId,
    workflowDagResultId,
    workflowDagWithLoadingStatus,
    workflowDagResultWithLoadingStatus,
  } = useWorkflow(
    user.apiKey,
    workflowIdProp,
    workflowDagIdProp,
    workflowDagResultIdProp
  );

  const { breadcrumbs, operator } = useOpeartor(
    operatorIdProp,
    wfBreadcrumbs,
    workflowDagWithLoadingStatus,
    workflowDagResultWithLoadingStatus,
    !sideSheetMode,
    'Check'
  );

  const artifactId = operator?.outputs[0];
  const artifactHistoryWithLoadingStatus = useArtifactHistory(
    user.apiKey,
    artifactId,
    workflowId,
    workflowDagResultWithLoadingStatus
  );

  return (
    <Layout breadcrumbs={breadcrumbs} user={user}>
      <RequireDagOrResult
        dagWithLoadingStatus={workflowDagWithLoadingStatus}
        dagResultWithLoadingStatus={workflowDagResultWithLoadingStatus}
      >
        <RequireOperator operator={operator}>
          <WithOperatorHeader
            workflowId={workflowId}
            dagId={workflowDagId}
            dagResultId={workflowDagResultId}
            dagWithLoadingStatus={workflowDagWithLoadingStatus}
            dagResultWithLoadingStatus={workflowDagResultWithLoadingStatus}
            operator={operator}
            sideSheetMode={sideSheetMode}
          >
            {workflowDagResultWithLoadingStatus && (
              <Box
                width={sideSheetMode ? 'auto' : '49.2%'}
                marginTop={sideSheetMode ? '16px' : '40px'}
              >
                <CheckHistory
                  historyWithLoadingStatus={artifactHistoryWithLoadingStatus}
                  checkLevel={operator?.spec?.check?.level}
                />
              </Box>
            )}
          </WithOperatorHeader>
        </RequireOperator>
      </RequireDagOrResult>
    </Layout>
  );
};

export default CheckDetailsPage;
