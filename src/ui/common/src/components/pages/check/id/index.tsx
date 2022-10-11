import { faChevronRight } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { CircularProgress, Link } from '@mui/material';
import Accordion from '@mui/material/Accordion';
import AccordionDetails from '@mui/material/AccordionDetails';
import AccordionSummary from '@mui/material/AccordionSummary';
import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link as RouterLink, useLocation, useParams } from 'react-router-dom';

import { BreadcrumbLink } from '../../../../components/layouts/NavBar';
import { handleGetWorkflowDagResult } from '../../../../handlers/getWorkflowDagResult';
import { handleListArtifactResults } from '../../../../handlers/listArtifactResults';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import { Data } from '../../../../utils/data';
import { getPathPrefix } from '../../../../utils/getPathPrefix';
import { isFailed, isInitial, isLoading } from '../../../../utils/shared';
import DefaultLayout from '../../../layouts/default';
import CheckTableItem from '../../../tables/CheckTableItem';
import CheckHistory from '../../../workflows/artifact/check/history';
import DetailsPageHeader from '../../components/DetailsPageHeader';
import { LayoutProps } from '../../types';

type CheckDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
  workflowIdProp?: string;
  workflowDagResultIdProp?: string;
  operatorIdProp?: string;
  sideSheetMode?: boolean;
};

const CheckDetailsPage: React.FC<CheckDetailsPageProps> = ({
  user,
  Layout = DefaultLayout,
  workflowIdProp,
  workflowDagResultIdProp,
  operatorIdProp,
  sideSheetMode = false,
}) => {
  const dispatch: AppDispatch = useDispatch();
  let { workflowId, workflowDagResultId, checkOperatorId } = useParams();
  const path = useLocation().pathname;

  if (workflowIdProp) {
    workflowId = workflowIdProp;
  }

  if (workflowDagResultIdProp) {
    workflowDagResultId = workflowDagResultIdProp;
  }

  if (operatorIdProp) {
    checkOperatorId = operatorIdProp;
  }

  const [metricsExpanded, setMetricsExpanded] = useState<boolean>(true);
  const [artifactsExpanded, setArtifactsExpanded] = useState<boolean>(true);

  const workflowDagResultWithLoadingStatus = useSelector(
    (state: RootState) =>
      state.workflowDagResultsReducer.results[workflowDagResultId]
  );

  const workflow = useSelector((state: RootState) => state.workflowReducer);

  const operator = (workflowDagResultWithLoadingStatus?.result?.operators ??
    {})[checkOperatorId];

  const pathPrefix = getPathPrefix();
  const workflowLink = `${pathPrefix}/workflow/${workflowId}?workflowDagResultId=${workflowDagResultId}`;
  const breadcrumbs = [
    BreadcrumbLink.HOME,
    BreadcrumbLink.WORKFLOWS,
    new BreadcrumbLink(workflowLink, workflow.selectedDag.metadata.name),
    new BreadcrumbLink(path, operator ? operator.name : 'Check'),
  ];

  const artifactId = operator?.outputs[0];

  const artifactHistoryWithLoadingStatus = useSelector((state: RootState) =>
    !!artifactId
      ? state.artifactResultsReducer.artifacts[artifactId]
      : undefined
  );

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
      // Queue up the artifacts historical results for loading.
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
    if (!!operator && !sideSheetMode) {
      document.title = `${operator.name} | Aqueduct`;
    }
  }, [operator]);

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

  const mockCheckArtifactData = [
    // severity in the mock corresponnds with level in the API response.
    {
      status: 'succeeded',
      level: 'warning',
      result: 'True',
      date_completed: '3/14/2022 4:00 PST',
    },
    {
      status: 'succeeded',
      level: 'warning',
      result: 'False',
      date_completed: '3/14/2022 4:00 PST',
    },
    {
      status: 'succeeded',
      level: 'error',
      result: 'True',
      date_completed: '3/14/2022 4:00 PST',
    },
  ];
  const historicalCheckData: Data = {
    schema: {
      fields: [
        { name: 'status', type: 'varchar' },
        { name: 'level', type: 'varchar' },
        { name: 'result', type: 'varchar' },
        { name: 'date_completed', type: 'varchar' },
      ],
      pandas_version: '0.0.1', // Not sure what actual value to put here, just filling in for now :)
    },
    data: mockCheckArtifactData,
  };

  // Function to get the numerical value of the metric output
  // TODO: Use this inside of the accordion component below.
  // NOTE: This code is shared with the metric details page, perhaps we should make this into a hook or component.
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
      <Box key={artifactId} display="flex">
        <CheckTableItem checkValue={artifactResult.result.content_serialized} />
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
  });

  return (
    <Layout breadcrumbs={breadcrumbs} user={user}>
      <Box width={'800px'}>
        {!sideSheetMode && (
          <Box width="100%">
            <DetailsPageHeader
              name={operator ? operator.name : 'Check Details'}
            />
            {operator?.description && (
              <Typography variant="body1">{operator.description}</Typography>
            )}
          </Box>
        )}

        <Box width="100%" marginTop={sideSheetMode ? '16px' : '40px'}>
          <Typography variant="h5" marginBottom="8px">
            Recent Results
          </Typography>
          <CheckHistory
            historyWithLoadingStatus={artifactHistoryWithLoadingStatus}
            checkLevel={operator?.spec?.check?.level}
          />
        </Box>
        {/* commenting out metrics for now as we figure out what to do with them */}
        {/* <Box width="100%" marginTop="32px">
                    <Typography variant="h5">Related Outputs</Typography>
                    <Accordion
                        expanded={metricsExpanded}
                        onChange={() => {
                            setMetricsExpanded(!metricsExpanded);
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
                                Metrics:
                            </Typography>
                        </AccordionSummary>
                        <AccordionDetails>
                            <React.Fragment>{operatorOutputsList}</React.Fragment>
                        </AccordionDetails>
                    </Accordion>
                </Box> */}

        <Box width="100%" marginTop="32px">
          <Accordion
            expanded={artifactsExpanded}
            onChange={() => {
              setArtifactsExpanded(!artifactsExpanded);
            }}
          >
            <AccordionSummary
              expandIcon={<FontAwesomeIcon icon={faChevronRight} />}
              sx={{
                '& .MuiAccordionSummary-expandIconWrapper.Mui-expanded': {
                  transform: 'rotate(90deg)',
                },
              }}
              aria-controls="artifacts-accordion-content"
              id="artifacts-accordion-header"
            >
              <Typography
                sx={{ width: '33%', flexShrink: 0 }}
                variant="h5"
                component="div"
                marginBottom="8px"
              >
                Artifacts:
              </Typography>
            </AccordionSummary>
            <AccordionDetails>
              <React.Fragment>{operatorOutputsList}</React.Fragment>
            </AccordionDetails>
          </Accordion>
        </Box>
      </Box>
    </Layout>
  );
};

export default CheckDetailsPage;
