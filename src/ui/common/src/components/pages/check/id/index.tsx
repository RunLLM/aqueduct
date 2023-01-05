import {
  faCircleExclamation,
  faTriangleExclamation,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { CircularProgress } from '@mui/material';
import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useLocation, useParams } from 'react-router-dom';

import { BreadcrumbLink } from '../../../../components/layouts/NavBar';
import { handleGetWorkflowDagResult } from '../../../../handlers/getWorkflowDagResult';
import { handleListArtifactResults } from '../../../../handlers/listArtifactResults';
import { AppDispatch, RootState } from '../../../../stores/store';
import { theme } from '../../../../styles/theme/theme';
import UserProfile from '../../../../utils/auth';
import { getPathPrefix } from '../../../../utils/getPathPrefix';
import { CheckLevel } from '../../../../utils/operators';
import { isFailed, isInitial, isLoading } from '../../../../utils/shared';
import DefaultLayout from '../../../layouts/default';
import CheckHistory from '../../../workflows/artifact/check/history';
import ArtifactSummaryList from '../../../workflows/artifact/summaryList';
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

  const workflowDagResultWithLoadingStatus = useSelector(
    (state: RootState) =>
      state.workflowDagResultsReducer.results[workflowDagResultId]
  );

  const operator = (workflowDagResultWithLoadingStatus?.result?.operators ??
    {})[checkOperatorId];

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
    new BreadcrumbLink(
      workflowLink,
      workflowDagResultWithLoadingStatus?.result?.name ?? 'Workflow'
    ),
    new BreadcrumbLink(path, operator ? operator.name : 'Check'),
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
      document.title = `${operator.name} | Aqueduct`;
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
          <AlertTitle>Failed to load workflow.</AlertTitle>
          {workflowDagResultWithLoadingStatus.status.err}
        </Alert>
      </Layout>
    );
  }

  const mapArtifacts = (artfIds: string[]) =>
    artfIds
      .map((artifactId) => {
        // We do a structuredClone so that we can modify this -- otherwise, it's an unmodifiable pointer to a
        // Redux object.
        const artifactResult = structuredClone(
          (workflowDagResultWithLoadingStatus.result?.artifacts ?? {})[
            artifactId
          ]
        );

        if (!artifactResult) {
          return artifactResult;
        }

        const operatorType =
          workflowDagResultWithLoadingStatus.result?.operators[
            artifactResult.from
          ]?.spec.type;
        artifactResult.operatorType = operatorType;

        return artifactResult;
      })
      .filter((artf) => !!artf);
  const inputs = operator ? mapArtifacts(operator.inputs) : [];
  const outputs = operator ? mapArtifacts(operator.outputs) : [];

  const checkLevel = operator ? operator.spec.check.level : 'Check Level';
  const checkLevelDisplay = (
    <Box sx={{ display: 'flex', alignItems: 'center' }} mb={2}>
      <Typography variant="body2" sx={{ color: 'gray.800' }}>
        Check Level
      </Typography>
      <Typography variant="body1" sx={{ mx: 1 }}>
        {checkLevel.charAt(0).toUpperCase() + checkLevel.slice(1)}
      </Typography>
      <FontAwesomeIcon
        icon={
          checkLevel === CheckLevel.Error
            ? faCircleExclamation
            : faTriangleExclamation
        }
        color={
          checkLevel === CheckLevel.Error
            ? theme.palette.red[600]
            : theme.palette.orange[600]
        }
      />
    </Box>
  );

  return (
    <Layout breadcrumbs={breadcrumbs} user={user}>
      <Box width={sideSheetMode ? 'auto' : 'auto'}>
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

        <Box width="100%" paddingTop={sideSheetMode ? '16px' : '24px'}>
          {checkLevelDisplay}

          <Box display="flex" width="100%">
            <Box width="100%" sx={{ mr: '32px' }}>
              <ArtifactSummaryList
                title={'Inputs'}
                workflowId={workflowId}
                dagResultId={workflowDagResultId}
                artifactResults={inputs}
                collapsePrimitives={false}
              />
            </Box>
            <Box width="100%">
              <ArtifactSummaryList
                title={'Outputs'}
                workflowId={workflowId}
                dagResultId={workflowDagResultId}
                artifactResults={outputs}
              />
            </Box>
          </Box>
        </Box>

        <Divider sx={{ my: '32px' }} />

        <Box
          width={sideSheetMode ? 'auto' : '49.2%'}
          marginTop={sideSheetMode ? '16px' : '40px'}
        >
          <CheckHistory
            historyWithLoadingStatus={artifactHistoryWithLoadingStatus}
            checkLevel={operator?.spec?.check?.level}
          />
        </Box>
      </Box>
    </Layout>
  );
};

export default CheckDetailsPage;
