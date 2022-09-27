import { CircularProgress } from '@mui/material';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
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

  const [inputsExpanded, setInputsExpanded] = useState<boolean>(true);
  const [outputsExpanded, setOutputsExpanded] = useState<boolean>(true);

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

  const listStyle = {
    width: '100%',
    maxWidth: 360,
    bgcolor: 'background.paper',
  };

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

  const inputs = operator.inputs
    .map(
      (artifactId) =>
        (workflowDagResultWithLoadingStatus.result?.artifacts ?? {})[artifactId]
    )
    .filter((artf) => !!artf);
  const outputs = operator.outputs
    .map(
      (artifactId) =>
        (workflowDagResultWithLoadingStatus.result?.artifacts ?? {})[artifactId]
    )
    .filter((artf) => !!artf);

  return (
    <Layout user={user}>
      <Box width={'800px'}>
        <Box width="100%">
          <Box width="100%">
            <DetailsPageHeader name={operator.name} />
            {operator.description && (
              <Typography variant="body1">{operator.description}</Typography>
            )}
          </Box>

          <Box display="flex" width="100%" paddingTop="40px">
            <Box width="100%">
              <ArtifactSummaryList
                title={'Inputs:'}
                workflowId={workflowId}
                dagResultId={workflowDagResultId}
                artifactResults={inputs}
                initiallyExpanded={true}
              />
            </Box>
            <Box width="32px" />
            <Box width="100%">
              <ArtifactSummaryList
                title={'Outputs:'}
                workflowId={workflowId}
                dagResultId={workflowDagResultId}
                artifactResults={outputs}
                initiallyExpanded={true}
              />
            </Box>
          </Box>

          <Box width="100%" marginTop="12px">
            <Typography variant="h5" component="div" marginBottom="8px">
              Historical Outputs:
            </Typography>
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
