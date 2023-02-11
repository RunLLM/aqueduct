import { CircularProgress, Divider } from '@mui/material';
import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useLocation, useParams } from 'react-router-dom';

import { BreadcrumbLink } from '../../../../components/layouts/NavBar';
import { handleGetArtifactResultContent } from '../../../../handlers/getArtifactResultContent';
import { handleGetWorkflowDagResult } from '../../../../handlers/getWorkflowDagResult';
import { getMetricsAndChecksOnArtifact } from '../../../../handlers/responses/dag';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import { getPathPrefix } from '../../../../utils/getPathPrefix';
import { OperatorType } from '../../../../utils/operators';
import ExecutionStatus, {
  isFailed,
  isInitial,
  isLoading,
  isSucceeded,
} from '../../../../utils/shared';
import DefaultLayout, { SidesheetContentWidth } from '../../../layouts/default';
import ArtifactContent from '../../../workflows/artifact/content';
import CsvExporter from '../../../workflows/artifact/csvExporter';
import {
  ChecksOverview,
  MetricsOverview,
} from '../../../workflows/artifact/metricsAndChecksOverview';
import OperatorSummaryList from '../../../workflows/operator/summaryList';
import DetailsPageHeader from '../../components/DetailsPageHeader';
import { LayoutProps } from '../../types';

type ArtifactDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
  workflowIdProp?: string;
  workflowDagResultIdProp?: string;
  operatorIdProp?: string;
  sideSheetMode?: boolean;
};

const ArtifactDetailsPage: React.FC<ArtifactDetailsPageProps> = ({
  user,
  Layout = DefaultLayout,
  workflowIdProp,
  workflowDagResultIdProp,
  operatorIdProp,
  sideSheetMode = false,
}) => {
  const dispatch: AppDispatch = useDispatch();
  let { workflowId, workflowDagResultId, artifactId } = useParams();
  const path = useLocation().pathname;

  if (workflowIdProp) {
    workflowId = workflowIdProp;
  }

  if (workflowDagResultIdProp) {
    workflowDagResultId = workflowDagResultIdProp;
  }

  if (operatorIdProp) {
    artifactId = operatorIdProp;
  }

  const workflowDagResultWithLoadingStatus = useSelector(
    (state: RootState) =>
      state.workflowDagResultsReducer.results[workflowDagResultId]
  );

  const artifactContents = useSelector(
    (state: RootState) => state.artifactResultContentsReducer.contents
  );

  const artifact = (workflowDagResultWithLoadingStatus?.result?.artifacts ??
    {})[artifactId];

  const artifactResultId = artifact?.result?.id;
  const contentWithLoadingStatus = artifactResultId
    ? artifactContents[artifactResultId]
    : undefined;

  const { metrics, checks } =
    !!workflowDagResultWithLoadingStatus &&
    isSucceeded(workflowDagResultWithLoadingStatus.status)
      ? getMetricsAndChecksOnArtifact(
          workflowDagResultWithLoadingStatus?.result,
          artifactId
        )
      : { metrics: [], checks: [] };

  const pathPrefix = getPathPrefix();
  const workflowLink = `${pathPrefix}/workflow/${workflowId}?workflowDagResultId=${workflowDagResultId}`;
  const breadcrumbs = [
    BreadcrumbLink.HOME,
    BreadcrumbLink.WORKFLOWS,
    new BreadcrumbLink(
      workflowLink,
      workflowDagResultWithLoadingStatus?.result?.name ?? 'Workflow'
    ),
    new BreadcrumbLink(path, artifact ? artifact.name : 'Artifact'),
  ];

  useEffect(() => {
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
    user.apiKey,
    workflowDagResultId,
    workflowDagResultWithLoadingStatus,
    workflowId,
  ]);

  useEffect(() => {
    if (!!artifact) {
      if (!sideSheetMode) {
        document.title = `${
          artifact ? artifact.name : 'Artifact Details'
        } | Aqueduct`;
      }

      if (
        !!artifact.result &&
        // intentional '==' to check undefined or null.
        artifact.result.content_serialized == null &&
        !contentWithLoadingStatus
      ) {
        dispatch(
          handleGetArtifactResultContent({
            apiKey: user.apiKey,
            artifactId,
            artifactResultId,
            workflowDagResultId,
          })
        );
      }
    }
  }, [
    artifact,
    artifactId,
    artifactResultId,
    contentWithLoadingStatus,
    dispatch,
    sideSheetMode,
    user.apiKey,
    workflowDagResultId,
  ]);

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
          <AlertTitle>Failed to load workflow.</AlertTitle>
          {workflowDagResultWithLoadingStatus.status.err}
        </Alert>
      </Layout>
    );
  }

  if (!artifact) {
    return (
      <Layout breadcrumbs={breadcrumbs} user={user}>
        <Alert severity="error">
          <AlertTitle>Failed to load artifact.</AlertTitle>
          Artifact {artifactId} does not exist on this workflow.
        </Alert>
      </Layout>
    );
  }

  const mapOperators = (opIds: string[]) =>
    opIds
      .map(
        (opId) =>
          (workflowDagResultWithLoadingStatus.result?.operators ?? {})[opId]
      )
      .filter((op) => !!op && op.spec?.type !== OperatorType.Param);

  const inputs = mapOperators([artifact.from]);
  const outputs = mapOperators(artifact.to ? artifact.to : []);

  let upstream_pending = false;
  inputs.some((operator) => {
    const operator_pending =
      operator.result.exec_state.status === ExecutionStatus.Pending;
    if (operator_pending) {
      upstream_pending = operator_pending;
    }
    return operator_pending;
  });

  const artifactStatus = artifact?.result?.exec_state?.status;
  const previewAvailable =
    artifactStatus && artifactStatus !== ExecutionStatus.Canceled;

  let preview = (
    <>
      <Divider sx={{ marginY: '32px' }} />

      <Box marginBottom="32px">
        <Alert severity="warning">
          An upstream operator failed, causing this artifact to not be created.
        </Alert>
      </Box>
    </>
  );

  if (upstream_pending) {
    preview = (
      <>
        <Divider sx={{ marginY: '32px' }} />

        <Box marginBottom="32px">
          <Alert severity="warning">
            An upstream operator is in progress so this artifact is not yet
            created.
          </Alert>
        </Box>
      </>
    );
  } else if (previewAvailable) {
    preview = (
      <>
        <Divider sx={{ marginY: '32px' }} />
        <Box width="100%" marginTop="12px">
          <Typography
            variant="h6"
            component="div"
            marginBottom="8px"
            fontWeight="normal"
          >
            Preview
          </Typography>
          <ArtifactContent
            artifact={artifact}
            contentWithLoadingStatus={contentWithLoadingStatus}
          />
        </Box>

        <Divider sx={{ marginY: '32px' }} />
      </>
    );
  }

  return (
    <Layout breadcrumbs={breadcrumbs} user={user}>
      <Box width={sideSheetMode ? SidesheetContentWidth : '100%'}>
        <Box width="100%">
          {!sideSheetMode && (
            <Box width="100%" display="flex" alignItems="center">
              <DetailsPageHeader
                name={artifact ? artifact.name : 'Artifact'}
                status={artifactStatus}
              />
              <CsvExporter
                artifact={artifact}
                contentWithLoadingStatus={contentWithLoadingStatus}
              />
            </Box>
          )}

          <Box display="flex" width="100%" mt={sideSheetMode ? '16px' : '64px'}>
            {inputs.length > 0 && (
              <Box width="100%" mr="32px">
                <OperatorSummaryList
                  title={'Generated By'}
                  workflowId={workflowId}
                  dagResultId={workflowDagResultId}
                  operatorResults={inputs}
                />
              </Box>
            )}

            {outputs.length > 0 && (
              <Box width="100%">
                <OperatorSummaryList
                  title={'Consumed By'}
                  workflowId={workflowId}
                  dagResultId={workflowDagResultId}
                  operatorResults={outputs}
                />
              </Box>
            )}
          </Box>

          {preview}

          <Box display="flex" width="100%">
            <MetricsOverview metrics={metrics} />
            <Box width="96px" />
            <ChecksOverview checks={checks} />
          </Box>
        </Box>
      </Box>
    </Layout>
  );
};

export default ArtifactDetailsPage;
