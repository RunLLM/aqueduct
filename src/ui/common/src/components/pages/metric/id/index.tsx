import { CircularProgress, Divider } from '@mui/material';
import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useLocation, useParams } from 'react-router-dom';

import { BreadcrumbLink } from '../../../../components/layouts/NavBar';
import { handleGetWorkflowDagResult } from '../../../../handlers/getWorkflowDagResult';
import { handleListArtifactResults } from '../../../../handlers/listArtifactResults';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import { getPathPrefix } from '../../../../utils/getPathPrefix';
import { isFailed, isInitial, isLoading } from '../../../../utils/shared';
import DefaultLayout from '../../../layouts/default';
import MetricsHistory from '../../../workflows/artifact/metric/history';
import ArtifactSummaryList from '../../../workflows/artifact/summaryList';
import DetailsPageHeader from '../../components/DetailsPageHeader';
import { LayoutProps } from '../../types';

type MetricDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
  workflowIdProp?: string;
  workflowDagResultIdProp?: string;
  operatorIdProp?: string;
  // true if shown as a sidesheet instead of a page.
  sideSheetMode?: boolean;
};

const MetricDetailsPage: React.FC<MetricDetailsPageProps> = ({
  user,
  Layout = DefaultLayout,
  workflowIdProp,
  workflowDagResultIdProp,
  operatorIdProp,
  sideSheetMode = false,
}) => {
  const dispatch: AppDispatch = useDispatch();
  let { workflowId, workflowDagResultId, metricOperatorId } = useParams();
  const path = useLocation().pathname;

  if (workflowIdProp) {
    workflowId = workflowIdProp;
  }

  if (workflowDagResultIdProp) {
    workflowDagResultId = workflowDagResultIdProp;
  }

  if (operatorIdProp) {
    metricOperatorId = operatorIdProp;
  }

  const workflowDagResultWithLoadingStatus = useSelector(
    (state: RootState) =>
      state.workflowDagResultsReducer.results[workflowDagResultId]
  );

  const workflow = useSelector((state: RootState) => state.workflowReducer);

  const operator = (workflowDagResultWithLoadingStatus?.result?.operators ??
    {})[metricOperatorId];

  const artifactId = operator?.outputs[0];
  const artifactHistoryWithLoadingStatus = useSelector((state: RootState) =>
    !!artifactId
      ? state.artifactResultsReducer.artifacts[artifactId]
      : undefined
  );

  const pathPrefix = getPathPrefix();
  const workflowLink = `${pathPrefix}/workflow/${workflowId}?workflowDagResultId=${workflowDagResultId}`;
  const breadcrumbs = [
    BreadcrumbLink.HOME,
    BreadcrumbLink.WORKFLOWS,
    new BreadcrumbLink(workflowLink, workflow.selectedDag?.metadata.name),
    new BreadcrumbLink(path, operator ? operator.name : 'Metric'),
  ];

  useEffect(() => {
    if (!sideSheetMode) {
      document.title = 'Metric Details | Aqueduct';
    }

    // Load workflow dag result if it's not cached
    if (
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
    sideSheetMode,
    user.apiKey,
    workflowDagResultId,
    workflowDagResultWithLoadingStatus,
    workflowId,
  ]);

  useEffect(() => {
    // Load artifact history once workflow dag results finished loading
    // and the result is not cached
    if (
      !artifactHistoryWithLoadingStatus &&
      !!artifactId &&
      !!workflowDagResultWithLoadingStatus &&
      !isInitial(workflowDagResultWithLoadingStatus.status) &&
      !isLoading(workflowDagResultWithLoadingStatus.status)
    ) {
      dispatch(
        handleListArtifactResults({
          apiKey: user.apiKey,
          workflowId,
          artifactId,
        })
      );
    }
  }, [
    workflowDagResultWithLoadingStatus,
    artifactId,
    artifactHistoryWithLoadingStatus,
    dispatch,
    user.apiKey,
    workflowId,
  ]);

  useEffect(() => {
    if (!!operator && !sideSheetMode) {
      // this should only be set when the user is viewing this as a full page, not side sheet.
      document.title = `${
        operator ? operator.name : 'Operator Details'
      } | Aqueduct`;
    }
  }, [operator, sideSheetMode]);

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

  if (isFailed(workflowDagResultWithLoadingStatus.status)) {
    return (
      <Layout breadcrumbs={breadcrumbs} user={user}>
        <Alert severity="error">
          <AlertTitle>Failed to load workflow</AlertTitle>
          {workflowDagResultWithLoadingStatus.status.err}
        </Alert>
      </Layout>
    );
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
      <Box width={sideSheetMode ? 'auto' : 'auto'}>
        <Box width="100%" mb={3}>
          {!sideSheetMode && (
            <Box width="100%">
              <DetailsPageHeader name={operator ? operator.name : 'Operator'} />
              {operator.description && (
                <Typography variant="body1">{operator.description}</Typography>
              )}
            </Box>
          )}
          <Box
            display="flex"
            width="100%"
            paddingTop={sideSheetMode ? '16px' : '40px'}
          >
            <Box width="100%">
              <ArtifactSummaryList
                title={'Inputs'}
                workflowId={workflowId}
                dagResultId={workflowDagResultId}
                artifactResults={inputs}
                appearance="value"
              />
            </Box>
            <Box width="32px" />
            <Box width="100%">
              <ArtifactSummaryList
                title={'Outputs'}
                workflowId={workflowId}
                dagResultId={workflowDagResultId}
                artifactResults={outputs}
                appearance="value"
              />
            </Box>
          </Box>

          <Divider sx={{ my: '32px' }} />

          <Box
            width={sideSheetMode ? 'auto' : '49.2%'}
            marginTop={sideSheetMode ? '16px' : '40px'}
          >
            <MetricsHistory
              historyWithLoadingStatus={artifactHistoryWithLoadingStatus}
            />
          </Box>
        </Box>
      </Box>
    </Layout>
  );
};

export default MetricDetailsPage;
