import { Divider } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import DefaultLayout from '../../../../components/layouts/default';
import LogViewer from '../../../../components/LogViewer';
import WithOperatorHeader from '../../../../components/operators/WithOperatorHeader';
import RequireOperator from '../../../operators/RequireOperator';
import OperatorSpecDetails from '../../../workflows/operator/specDetails';
import RequireDagOrResult from '../../../workflows/RequireDagOrResult';
import { LayoutProps } from '../../types';
import useWorkflow from '../../workflow/id/hook';
import useOpeartor from './hook';

type OperatorDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
  workflowIdProp?: string;
  workflowDagIdProp?: string;
  workflowDagResultIdProp?: string;
  operatorIdProp?: string;
  sideSheetMode?: boolean;
};

// Checked with file size=313285391 and handles that smoothly once loaded. However, takes a while to load.
const OperatorDetailsPage: React.FC<OperatorDetailsPageProps> = ({
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
    !sideSheetMode
  );

  const logs = operator?.result?.exec_state?.user_logs ?? {};
  const operatorError = operator?.result?.exec_state?.error;

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
            <Box>
              <Typography variant="h6" fontWeight="normal">
                Logs
              </Typography>
              {logs !== {} && <LogViewer logs={logs} err={operatorError} />}
            </Box>

            <Divider sx={{ my: '32px' }} />
            <OperatorSpecDetails user={user} operator={operator} />
          </WithOperatorHeader>
        </RequireOperator>
      </RequireDagOrResult>
    </Layout>
  );
};

export default OperatorDetailsPage;
