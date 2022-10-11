import { CircularProgress, Divider } from '@mui/material';
import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useLocation, useParams } from 'react-router-dom';

import { BreadcrumbLinks } from '../../../../components/layouts/NavBar';
import { handleGetArtifactResultContent } from '../../../../handlers/getArtifactResultContent';
import { handleGetWorkflowDagResult } from '../../../../handlers/getWorkflowDagResult';
import { getMetricsAndChecksOnArtifact } from '../../../../handlers/responses/dag';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import {
  isFailed,
  isInitial,
  isLoading,
  isSucceeded,
} from '../../../../utils/shared';
import DefaultLayout from '../../../layouts/default';
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

  const workflow = useSelector((state: RootState) => state.workflowReducer);

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

  useEffect(() => {
    document.title = 'Artifact Details | Aqueduct';

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
  }, []);

  useEffect(() => {
    if (!!artifact && !sideSheetMode) {
      document.title = `${
        artifact ? artifact.name : 'Artifact Details'
      } | Aqueduct`;

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
  }, [artifact]);

  if (
    !workflowDagResultWithLoadingStatus ||
    isInitial(workflowDagResultWithLoadingStatus.status) ||
    isLoading(workflowDagResultWithLoadingStatus.status)
  ) {
    return (
      <Layout
        breadcrumbs={[
          BreadcrumbLinks.HOME,
          BreadcrumbLinks.WORKFLOWS,
          new BreadcrumbLinks(
            path.split('/artifact/')[0],
            workflow.selectedDag.metadata.name
          ),
          new BreadcrumbLinks(path, artifact ? artifact.name : 'Artifact'),
        ]}
        user={user}
      >
        <CircularProgress />
      </Layout>
    );
  }

  if (isFailed(workflowDagResultWithLoadingStatus.status)) {
    return (
      <Layout
        breadcrumbs={[
          BreadcrumbLinks.HOME,
          BreadcrumbLinks.WORKFLOWS,
          new BreadcrumbLinks(
            path.split('/artifact/')[0],
            workflow.selectedDag.metadata.name
          ),
          new BreadcrumbLinks(path, artifact ? artifact.name : 'Artifact'),
        ]}
        user={user}
      >
        <Alert severity="error">
          <AlertTitle>Failed to load workflow.</AlertTitle>
          {workflowDagResultWithLoadingStatus.status.err}
        </Alert>
      </Layout>
    );
  }

  if (!artifact) {
    return (
      <Layout
        breadcrumbs={[
          BreadcrumbLinks.HOME,
          BreadcrumbLinks.WORKFLOWS,
          new BreadcrumbLinks(
            path.split('/artifact/')[0],
            workflow.selectedDag.metadata.name
          ),
          new BreadcrumbLinks(path, artifact ? artifact.name : 'Artifact'),
        ]}
        user={user}
      >
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
      .filter((op) => !!op);

  const inputs = mapOperators([artifact.from]);
  const outputs = mapOperators(artifact.to ? artifact.to : []);

  return (
    <Layout
      breadcrumbs={[
        BreadcrumbLinks.HOME,
        BreadcrumbLinks.WORKFLOWS,
        new BreadcrumbLinks(
          path.split('/artifact/')[0],
          workflow.selectedDag.metadata.name
        ),
        new BreadcrumbLinks(path, artifact ? artifact.name : 'Artifact'),
      ]}
      user={user}
    >
      <Box width={'800px'}>
        <Box width="100%">
          {!sideSheetMode && (
            <Box width="100%" display="flex" alignItems="center">
              <DetailsPageHeader name={artifact ? artifact.name : 'Artifact'} />
              <CsvExporter
                artifact={artifact}
                contentWithLoadingStatus={contentWithLoadingStatus}
              />
            </Box>
          )}

          <Box display="flex" width="100%" mt={sideSheetMode ? '16px' : '64px'}>
            <Box width="100%" mr="32px">
              <OperatorSummaryList
                title={'Generated By'}
                workflowId={workflowId}
                dagResultId={workflowDagResultId}
                operatorResults={inputs}
                initiallyExpanded={true}
              />
            </Box>

            <Box width="100%">
              <OperatorSummaryList
                title={'Consumed By'}
                workflowId={workflowId}
                dagResultId={workflowDagResultId}
                operatorResults={outputs}
                initiallyExpanded={true}
              />
            </Box>
          </Box>

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
