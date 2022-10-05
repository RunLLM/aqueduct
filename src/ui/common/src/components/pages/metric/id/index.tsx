import { CircularProgress, Divider } from '@mui/material';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useParams } from 'react-router-dom';

import { handleGetWorkflowDagResult } from '../../../../handlers/getWorkflowDagResult';
import { handleListArtifactResults } from '../../../../handlers/listArtifactResults';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import { isFailed, isInitial, isLoading } from '../../../../utils/shared';
import DefaultLayout from '../../../layouts/default';
import MetricsHistory from '../../../workflows/artifact/metric/history';
import ArtifactSummaryList from '../../../workflows/artifact/summaryList';
import DetailsPageHeader from '../../components/DetailsPageHeader';
import { LayoutProps } from '../../types';

type MetricDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const MetricDetailsPage: React.FC<MetricDetailsPageProps> = ({
  user,
  Layout = DefaultLayout,
}) => {
  const dispatch: AppDispatch = useDispatch();
  const { workflowId, workflowDagResultId, metricOperatorId } = useParams();

  const workflowDagResultWithLoadingStatus = useSelector(
    (state: RootState) =>
      state.workflowDagResultsReducer.results[workflowDagResultId]
  );
  const operator = (workflowDagResultWithLoadingStatus?.result?.operators ??
    {})[metricOperatorId];
  const artifactId = operator?.outputs[0];
  const artifactHistoryWithLoadingStatus = useSelector((state: RootState) =>
    !!artifactId
      ? state.artifactResultsReducer.artifacts[artifactId]
      : undefined
  );

  useEffect(() => {
    document.title = 'Metric Details | Aqueduct';

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
  }, [workflowDagResultWithLoadingStatus, artifactId]);

  useEffect(() => {
    if (!!operator) {
      document.title = `${operator.name} | Aqueduct`;
    }
  }, [operator]);

  if (
    !workflowDagResultWithLoadingStatus ||
    isInitial(workflowDagResultWithLoadingStatus.status) ||
    isLoading(workflowDagResultWithLoadingStatus.status)
  ) {
    return (
      <Layout user={user}>
        <CircularProgress />
      </Layout>
    );
  }

  if (isFailed(workflowDagResultWithLoadingStatus.status)) {
    return (
      <Layout user={user}>
        <Alert title="Failed to load workflow">
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
    <Layout user={user}>
      <Box width={'800px'}>
        <Box width="100%" mb={3}>
          <Box width="100%">
            <DetailsPageHeader name={operator.name} />
            {operator.description && (
              <Typography variant="body1">{operator.description}</Typography>
            )}
          </Box>

          <Box display="flex" width="100%" paddingTop="40px">
            <Box width="100%">
              <ArtifactSummaryList
                title={'Inputs'}
                workflowId={workflowId}
                dagResultId={workflowDagResultId}
                artifactResults={inputs}
              />
            </Box>
            <Box width="32px" />
            <Box width="100%">
              <ArtifactSummaryList
                title={'Outputs'}
                workflowId={workflowId}
                dagResultId={workflowDagResultId}
                artifactResults={outputs}
              />
            </Box>
          </Box>

          <Divider sx={{ my: '32px' }} />

          <Box width="100%">
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
