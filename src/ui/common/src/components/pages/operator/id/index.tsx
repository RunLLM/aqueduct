import { CircularProgress, Divider } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useLocation, useNavigate, useParams } from 'react-router-dom';

import DefaultLayout, {
  SidesheetContentWidth,
} from '../../../../components/layouts/default';
import { BreadcrumbLink } from '../../../../components/layouts/NavBar';
import LogViewer from '../../../../components/LogViewer';
import { handleGetWorkflowDagResult } from '../../../../handlers/getWorkflowDagResult';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import { getPathPrefix } from '../../../../utils/getPathPrefix';
import {
  isInitial,
  isLoading,
  LoadingStatusEnum,
} from '../../../../utils/shared';
import ArtifactSummaryList from '../../../workflows/artifact/summaryList';
import OperatorSpecDetails from '../../../workflows/operator/specDetails';
import DetailsPageHeader from '../../components/DetailsPageHeader';
import { LayoutProps } from '../../types';

type OperatorDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
  workflowIdProp?: string;
  workflowDagResultIdProp?: string;
  operatorIdProp?: string;
  sideSheetMode?: boolean;
};

// Checked with file size=313285391 and handles that smoothly once loaded. However, takes a while to load.
const OperatorDetailsPage: React.FC<OperatorDetailsPageProps> = ({
  user,
  Layout = DefaultLayout,
  workflowIdProp,
  workflowDagResultIdProp,
  operatorIdProp,
  sideSheetMode = false,
}) => {
  const dispatch: AppDispatch = useDispatch();
  const navigate = useNavigate();
  let { workflowId, workflowDagResultId, operatorId } = useParams();
  const path = useLocation().pathname;

  if (workflowIdProp) {
    workflowId = workflowIdProp;
  }

  if (workflowDagResultIdProp) {
    workflowDagResultId = workflowDagResultIdProp;
  }

  if (operatorIdProp) {
    operatorId = operatorIdProp;
  }

  const workflowDagResultWithLoadingStatus = useSelector(
    (state: RootState) =>
      state.workflowDagResultsReducer.results[workflowDagResultId]
  );

  const operator = (workflowDagResultWithLoadingStatus?.result?.operators ??
    {})[operatorId];

  const pathPrefix = getPathPrefix();
  const workflowLink = `${pathPrefix}/workflow/${workflowId}?workflowDagResultId=${workflowDagResultId}`;
  const breadcrumbs = [
    BreadcrumbLink.HOME,
    BreadcrumbLink.WORKFLOWS,
    new BreadcrumbLink(
      workflowLink,
      workflowDagResultWithLoadingStatus?.result?.name ?? 'Workflow'
    ),
    new BreadcrumbLink(path, operator?.name || 'Operator'),
  ];

  useEffect(() => {
    if (
      // Load workflow dag result if it's not cached
      !workflowDagResultWithLoadingStatus ||
      isInitial(workflowDagResultWithLoadingStatus.status)
    ) {
      dispatch(
        handleGetWorkflowDagResult({
          apiKey: user.apiKey,
          workflowId,
          workflowDagResultId,
        })
      );
    }
  }, [
    dispatch,
    user.apiKey,
    workflowDagResultId,
    workflowDagResultWithLoadingStatus,
    workflowId,
  ]);

  useEffect(() => {
    if (!!operator && !sideSheetMode) {
      document.title = `${operator?.name || 'Operator'} | Aqueduct`;
    }
  }, [operator, sideSheetMode]);

  const logs = operator?.result?.exec_state?.user_logs ?? {};
  const operatorError = operator?.result?.exec_state?.error;

  const operatorStatus = operator?.result?.exec_state?.status;

  if (
    !workflowDagResultWithLoadingStatus ||
    isInitial(workflowDagResultWithLoadingStatus.status) ||
    isLoading(workflowDagResultWithLoadingStatus.status)
  ) {
    return (
      <Layout breadcrumbs={breadcrumbs} user={user}>
        <CircularProgress />
      </Layout>
    );
  }

  // This workflow doesn't exist.
  if (
    workflowDagResultWithLoadingStatus.status.loading ===
    LoadingStatusEnum.Failed
  ) {
    navigate('/404');
    return null;
  }

  const mapArtifacts = (artfIds: string[]) =>
    artfIds
      .map(
        (artifactId) =>
          (workflowDagResultWithLoadingStatus.result?.artifacts ?? {})[
          artifactId
          ]
      )
      .filter((artf) => !!artf);
  const inputs = mapArtifacts(operator.inputs);
  const outputs = mapArtifacts(operator.outputs);

  return (
    <Layout breadcrumbs={breadcrumbs} user={user}>
      <Box width={sideSheetMode ? SidesheetContentWidth : '100%'}>
        <Box width="100%">
          {!sideSheetMode && (
            <Box width="100%">
              <DetailsPageHeader
                name={operator ? operator.name : 'Operator'}
                status={operatorStatus}
              />
              {operator?.description && (
                <Typography variant="body1">{operator?.description}</Typography>
              )}
            </Box>
          )}
          <Box display="flex" width="100%" pt={sideSheetMode ? '16px' : '64px'}>
            {inputs.length > 0 && (
              <Box width="100%" mr="32px">
                <ArtifactSummaryList
                  title="Inputs"
                  workflowId={workflowId}
                  dagResultId={workflowDagResultId}
                  artifactResults={inputs}
                  appearance="link"
                />
              </Box>
            )}

            {outputs.length > 0 && (
              <Box width="100%">
                <ArtifactSummaryList
                  title="Outputs"
                  workflowId={workflowId}
                  dagResultId={workflowDagResultId}
                  artifactResults={outputs}
                  appearance="link"
                />
              </Box>
            )}
          </Box>

          <Divider sx={{ my: '32px' }} />

          <Box>
            <Typography variant="h6" fontWeight="normal">
              Logs
            </Typography>
            {logs !== {} && <LogViewer logs={logs} err={operatorError} />}
          </Box>

          <Divider sx={{ my: '32px' }} />
          <OperatorSpecDetails user={user} operator={operator} />
        </Box>
      </Box>
    </Layout>
  );
};

export default OperatorDetailsPage;
