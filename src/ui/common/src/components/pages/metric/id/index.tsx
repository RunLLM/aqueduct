import { faChevronRight } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { CircularProgress, Link, List, ListItem } from '@mui/material';
import Accordion from '@mui/material/Accordion';
import AccordionDetails from '@mui/material/AccordionDetails';
import AccordionSummary from '@mui/material/AccordionSummary';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
import Plot from 'react-plotly.js';
import { useDispatch, useSelector } from 'react-redux';
import { Link as RouterLink, useParams } from 'react-router-dom';

import PaginatedTable from '../../../../components/tables/PaginatedTable';
import { artifactTypeToIconMapping } from '../../../../components/workflows/nodes/nodeTypes';
import { handleGetWorkflowDagResult } from '../../../../handlers/getWorkflowDagResult';
import { handleListArtifactResults } from '../../../../handlers/listArtifactResults';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import { Data } from '../../../../utils/data';
import { getPathPrefix } from '../../../../utils/getPathPrefix';
import { isFailed, isInitial, isLoading } from '../../../../utils/shared';
import DefaultLayout from '../../../layouts/default';
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

  // Function to get the numerical value of the metric output
  const operatorOutputsList = operator.outputs.map((artifactId) => {
    const artifactResult = (workflowDagResultWithLoadingStatus.result
      ?.artifacts ?? {})[artifactId];
    if (!artifactResult) {
      return null;
    }

    if (
      !artifactResult.result ||
      artifactResult.result.content_serialized === undefined
    ) {
      // Link to appropriate artifact details page
      // Show tableIcon here as part of the link.
      return (
        <Box key={artifactId}>
          <Link
            to={`${getPathPrefix()}/workflow/${workflowId}/result/${workflowDagResultId}/artifact/${artifactId}`}
            component={RouterLink as any}
            sx={{ marginLeft: '16px' }}
            underline="none"
          >
            {artifactResult.name}
          </Link>
        </Box>
      );
    }

    return (
      <Box key={artifactId}>
        <Typography variant="body1">
          {artifactResult.result.content_serialized}
        </Typography>
      </Box>
    );
  });

  const operatorInputsList = operator.inputs.map((artifactId, index) => {
    const artifactResult = (workflowDagResultWithLoadingStatus.result
      ?.artifacts ?? {})[artifactId];
    if (!artifactResult) {
      return null;
    }

    return (
      <ListItem divider key={`metric-input-${index}`}>
        <Box display="flex">
          <Box
            sx={{
              width: '16px',
              height: '16px',
              color: 'rgba(0,0,0,0.54)',
            }}
          >
            <FontAwesomeIcon
              icon={artifactTypeToIconMapping[artifactResult.type]}
            />
          </Box>
          <Link
            to={`${getPathPrefix()}/workflow/${workflowId}/result/${workflowDagResultId}/artifact/${artifactId}`}
            component={RouterLink as any}
            sx={{ marginLeft: '16px' }}
            underline="none"
          >
            {artifactResult.name}
          </Link>
        </Box>
      </ListItem>
    );
  });

  let historicalOutputsSection = null;
  if (
    !artifactHistoryWithLoadingStatus ||
    isInitial(artifactHistoryWithLoadingStatus.status) ||
    isLoading(artifactHistoryWithLoadingStatus.status)
  ) {
    historicalOutputsSection = <CircularProgress />;
  } else if (isFailed(artifactHistoryWithLoadingStatus.status)) {
    historicalOutputsSection = (
      <Alert title="Failed to load historical data.">
        {artifactHistoryWithLoadingStatus.status.err}
      </Alert>
    );
  } else {
    const historicalData: Data = {
      schema: {
        fields: [
          { name: 'status', type: 'varchar' },
          { name: 'timestamp', type: 'varchar' },
          { name: 'value', type: 'float' },
        ],
        pandas_version: '0.0.1',
      },
      data: (artifactHistoryWithLoadingStatus.results?.results ?? []).map(
        (artifactStatusResult) => {
          return {
            status: artifactStatusResult.exec_state?.status ?? 'Unknown',
            timestamp: artifactStatusResult.exec_state?.timestamps?.finished_at,
            value: artifactStatusResult.content_serialized,
          };
        }
      ),
    };

    const dataToPlot = historicalData.data.filter(
      (x) => !!x['timestamp'] && !!x['value']
    );
    const timestamps = dataToPlot.map((x) => x['timestamp']);
    const values = dataToPlot.map((x) => x['value']);
    historicalOutputsSection = (
      <Box display="flex" justifyContent="center" flexDirection="column">
        <Plot
          data={[
            {
              x: timestamps,
              y: values,
              type: 'scatter',
              mode: 'lines+markers',
              marker: { color: 'red' },
            },
          ]}
          layout={{ width: '100%', height: '100%' }}
        />
        <PaginatedTable data={historicalData} />
      </Box>
    );
  }

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
              <Accordion
                expanded={inputsExpanded}
                onChange={() => {
                  setInputsExpanded(!inputsExpanded);
                }}
              >
                <AccordionSummary
                  expandIcon={<FontAwesomeIcon icon={faChevronRight} />}
                  sx={{
                    '& .MuiAccordionSummary-expandIconWrapper.Mui-expanded': {
                      transform: 'rotate(90deg)',
                    },
                  }}
                  aria-controls="input-accordion-content"
                  id="input-accordion-header"
                >
                  <Typography
                    sx={{ width: '33%', flexShrink: 0 }}
                    variant="h5"
                    component="div"
                    marginBottom="8px"
                  >
                    Inputs:
                  </Typography>
                </AccordionSummary>
                <AccordionDetails>
                  <List sx={listStyle}>{operatorInputsList}</List>
                </AccordionDetails>
              </Accordion>
            </Box>
            <Box width="32px" />
            <Box width="100%">
              <Accordion
                expanded={outputsExpanded}
                onChange={() => {
                  setOutputsExpanded(!outputsExpanded);
                }}
              >
                <AccordionSummary
                  expandIcon={<FontAwesomeIcon icon={faChevronRight} />}
                  sx={{
                    '& .MuiAccordionSummary-expandIconWrapper.Mui-expanded': {
                      transform: 'rotate(90deg)',
                    },
                  }}
                  aria-controls="panel1bh-content"
                  id="panel1bh-header"
                >
                  <Typography
                    sx={{ width: '33%', flexShrink: 0 }}
                    variant="h5"
                    component="div"
                    marginBottom="8px"
                  >
                    Output:
                  </Typography>
                </AccordionSummary>
                <AccordionDetails>
                  <React.Fragment>{operatorOutputsList}</React.Fragment>
                </AccordionDetails>
              </Accordion>
            </Box>
          </Box>

          <Box width="100%" marginTop="12px">
            <Typography variant="h5" component="div" marginBottom="8px">
              Historical Outputs:
            </Typography>
            {historicalOutputsSection}
          </Box>
        </Box>
      </Box>
    </Layout>
  );
};

export default MetricDetailsPage;
