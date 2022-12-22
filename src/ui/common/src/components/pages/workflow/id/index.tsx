import { faChevronRight } from '@fortawesome/free-solid-svg-icons';
import { faCircleDown } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Drawer } from '@mui/material';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import Typography from '@mui/material/Typography';
import { parse } from 'query-string';
import React, { useCallback, useEffect } from 'react';
import { ReactFlowProvider } from 'react-flow-renderer';
import { useDispatch, useSelector } from 'react-redux';
import { useLocation, useNavigate, useParams } from 'react-router-dom';

import { BreadcrumbLink } from '../../../../components/layouts/NavBar';
import { handleLoadIntegrations } from '../../../../reducers/integrations';
import {
  NodeType,
  resetSelectedNode,
} from '../../../../reducers/nodeSelection';
import {
  setAllSideSheetState,
  setBottomSideSheetOpenState,
} from '../../../../reducers/openSideSheet';
import {
  handleGetArtifactResults,
  handleGetOperatorResults,
  handleGetSelectDagPosition,
  handleGetWorkflow,
  resetState,
  selectResultIdx,
} from '../../../../reducers/workflow';
import { AppDispatch, RootState } from '../../../../stores/store';
import { theme } from '../../../../styles/theme/theme';
import UserProfile from '../../../../utils/auth';
import { Data } from '../../../../utils/data';
import { getPathPrefix } from '../../../../utils/getPathPrefix';
import { handleExportFunction } from '../../../../utils/operators';
import { exportCsv } from '../../../../utils/preview';
import {
  ExecutionStatus,
  LoadingStatusEnum,
  WidthTransition,
} from '../../../../utils/shared';
import {
  getDataSideSheetContent,
  sideSheetSwitcher,
} from '../../../../utils/sidesheets';
import DefaultLayout, {
  DefaultLayoutMargin,
  SidesheetButtonHeight,
  SidesheetWidth,
} from '../../../layouts/default';
import { Button } from '../../../primitives/Button.styles';
import ReactFlowCanvas from '../../../workflows/ReactFlowCanvas';
import WorkflowHeader, {
  WorkflowPageContentId,
} from '../../../workflows/workflowHeader';
import { LayoutProps } from '../../types';

type WorkflowPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const WorkflowPage: React.FC<WorkflowPageProps> = ({
  user,
  Layout = DefaultLayout,
}) => {
  const navigate = useNavigate();
  const dispatch: AppDispatch = useDispatch();
  const workflowId = useParams().id;
  const urlSearchParams = parse(window.location.search);
  const location = useLocation();
  const path = location.pathname;

  const currentNode = useSelector(
    (state: RootState) => state.nodeSelectionReducer.selected
  );
  const workflow = useSelector((state: RootState) => state.workflowReducer);
  const switchSideSheet = sideSheetSwitcher(dispatch);
  const artifactResult = useSelector(
    (state: RootState) => state.workflowReducer.artifactResults[currentNode.id]
  );

  const dagName = workflow.selectedDag?.metadata?.name;

  useEffect(() => {
    if (workflow.selectedDag !== undefined) {
      document.title = `${dagName} | Aqueduct`;
    }
  }, [workflow.selectedDag, dagName]);

  const resetWorkflowState = useCallback(() => {
    dispatch(resetState());
  }, [dispatch]);

  useEffect(() => {
    window.onpopstate = () => {
      resetWorkflowState();
    };

    resetWorkflowState();
  }, [resetWorkflowState]);

  useEffect(() => {
    if (
      workflow.selectedResult !== undefined &&
      !urlSearchParams.workflowDagResultId
    ) {
      navigate(
        `?workflowDagResultId=${encodeURI(workflow.selectedResult.id)}`,
        { replace: true }
      );
    }
  }, [workflow.selectedResult, urlSearchParams, navigate]);

  useEffect(() => {
    dispatch(handleGetWorkflow({ apiKey: user.apiKey, workflowId }));
    dispatch(handleLoadIntegrations({ apiKey: user.apiKey }));
  }, [dispatch, user.apiKey, workflowId]);

  useEffect(() => {
    if (workflow.dagResults && workflow.dagResults.length > 0) {
      let workflowDagResultIndex = 0;
      const { workflowDagResultId } = urlSearchParams;
      for (let i = 0; i < workflow.dagResults.length; i++) {
        if (workflow.dagResults[i].id === workflowDagResultId) {
          workflowDagResultIndex = i;
        }
      }
      if (workflowDagResultId !== workflow.selectedResult.id) {
        dispatch(setAllSideSheetState(false));
        // this is where selectedDag gets set
        dispatch(selectResultIdx(workflowDagResultIndex));
      }
    }
  }, [
    workflow.dagResults,
    urlSearchParams,
    workflow.selectedResult?.id,
    dispatch,
  ]);

  useEffect(() => {
    if (workflow.selectedDag) {
      dispatch(
        handleGetSelectDagPosition({
          apiKey: user.apiKey,
          operators: workflow.selectedDag?.operators,
          artifacts: workflow.selectedDag?.artifacts,
        })
      );
    }
  }, [dispatch, user.apiKey, workflow.selectedDag]);

  /**
   * This function dispatches calls to fetch artifact results and contents.
   *
   * This function is only activated when another similar fetch request
   * hasn't already been triggered.
   *
   * @param nodeId the UUID of the artifact for which we're retrieving
   * details.
   */

  const getArtifactResultDetails = useCallback(
    (nodeId: string) => {
      const artf = (workflow.selectedDag?.artifacts ?? {})[nodeId];
      if (!artf || !workflow.selectedResult) {
        return;
      }

      if (!(nodeId in workflow.artifactResults)) {
        dispatch(
          handleGetArtifactResults({
            apiKey: user.apiKey,
            workflowDagResultId: workflow.selectedResult.id,
            artifactId: nodeId,
          })
        );
      }
    },
    [
      dispatch,
      user.apiKey,
      workflow.artifactResults,
      workflow.selectedDag?.artifacts,
      workflow.selectedResult,
    ]
  );

  /**
   * This function fetches both the metadata of a particular operator as well
   * as the results of the artifacts that were both inputs and outputs for
   * this operator.
   *
   * This function is only activated when another similar fetch request
   * hasn't already been triggered.
   *
   * @param nodeId the UUID of an artifact for which we're retrieving
   * results.
   */
  const getOperatorResultDetails = useCallback(
    (nodeId: string) => {
      // Verify the node is indeed an operator, and a result is selected
      const op = (workflow.selectedDag?.operators ?? {})[nodeId];
      if (!op || !workflow.selectedResult) {
        return;
      }

      if (!(nodeId in workflow.operatorResults)) {
        dispatch(
          handleGetOperatorResults({
            apiKey: user.apiKey,
            workflowDagResultId: workflow.selectedResult.id,
            operatorId: nodeId,
          })
        );
      }

      for (const artfId of [...op.inputs, ...op.outputs]) {
        getArtifactResultDetails(artfId);
      }
    },
    [
      dispatch,
      getArtifactResultDetails,
      user.apiKey,
      workflow.operatorResults,
      workflow.selectedDag?.operators,
      workflow.selectedResult,
    ]
  );

  useEffect(() => {
    getOperatorResultDetails(currentNode.id);
    getArtifactResultDetails(currentNode.id);
  }, [
    currentNode?.id,
    getArtifactResultDetails,
    getOperatorResultDetails,
    workflow.selectedResult?.id,
  ]);

  const onPaneClicked = (event: React.MouseEvent) => {
    event.preventDefault();
    dispatch(setBottomSideSheetOpenState(false));

    // Reset selected node
    dispatch(resetSelectedNode());
  };

  const selectedDag = workflow.selectedDag;
  const getDagResultDetails = () => {
    if (
      workflow.loadingStatus.loading === LoadingStatusEnum.Succeeded &&
      !!selectedDag
    ) {
      for (const op of Object.values(selectedDag.operators)) {
        // We don't need to call getArtifactResultDetails because
        // getOperatorResultDetails automatically does that for us.
        getOperatorResultDetails(op.id);
      }
    }
  };

  useEffect(getDagResultDetails, [
    getOperatorResultDetails,
    selectedDag,
    workflow.loadingStatus.loading,
    workflow.selectedResult?.id,
  ]);

  // This workflow doesn't exist.
  if (workflow.loadingStatus.loading === LoadingStatusEnum.Failed) {
    navigate('/404');
    return null;
  }

  if (workflow.loadingStatus.loading !== LoadingStatusEnum.Succeeded) {
    return null;
  }

  // TODO: Remove openSideSheet reducer, as it's no longer used in the ui-redesign project
  // const sideSheetOpen = currentNode.type !== NodeType.None;

  const contentBottomOffsetInPx = `32px`;
  const getNodeLabel = () => {
    if (
      currentNode.type === NodeType.TableArtifact ||
      currentNode.type === NodeType.NumericArtifact ||
      currentNode.type === NodeType.BoolArtifact ||
      currentNode.type === NodeType.JsonArtifact ||
      currentNode.type === NodeType.StringArtifact ||
      currentNode.type === NodeType.ImageArtifact ||
      currentNode.type === NodeType.DictArtifact ||
      currentNode.type == NodeType.ListArtifact ||
      currentNode.type === NodeType.GenericArtifact
    ) {
      if (selectedDag.artifacts[currentNode.id]) {
        return selectedDag.artifacts[currentNode.id].name;
      }
      return 'Artifact Node';
    } else {
      if (selectedDag.operators[currentNode.id]) {
        return selectedDag.operators[currentNode.id].name;
      }
      return 'Operator Node';
    }
  };

  const getNodeActionButton = () => {
    const buttonStyle = {
      height: SidesheetButtonHeight,
      marginRight: '16px',
    };

    if (currentNode.type === NodeType.TableArtifact) {
      // Since workflow is pending, it doesn't have a result set yet.
      let artifactResultData: Data | null = null;
      if (
        artifactResult?.result &&
        artifactResult.result.exec_state.status === ExecutionStatus.Succeeded &&
        artifactResult.result.data.length > 0
      ) {
        artifactResultData = JSON.parse(artifactResult.result.data);
      }

      return (
        <Box>
          <Button
            style={buttonStyle}
            onClick={() => {
              // All we're really doing here is adding the artifactId onto the end of the URL.
              navigate(
                `${getPathPrefix()}/workflow/${workflowId}/result/${
                  workflow.selectedResult.id
                }/artifact/${currentNode.id}`
              );
            }}
          >
            View Artifact Details
          </Button>
          <Button
            style={buttonStyle}
            onClick={() =>
              exportCsv(artifactResultData, getNodeLabel().replaceAll(' ', '_'))
            }
          >
            Export CSV
          </Button>
        </Box>
      );
    }

    const operator = (workflow.selectedDag?.operators ?? {})[currentNode.id];
    const exportOpButton = (
      <Button
        style={{ ...buttonStyle, maxWidth: '300px' }}
        onClick={async () => {
          await handleExportFunction(
            user,
            currentNode.id,
            `${operator?.name ?? 'function'}.zip`
          );
        }}
        color="primary"
      >
        <FontAwesomeIcon icon={faCircleDown} />
        <Typography
          sx={{ ml: 1, overflow: 'hidden', textOverflow: 'ellipsis' }}
        >{`${operator?.name ?? 'function'}.zip`}</Typography>
      </Button>
    );

    if (currentNode.type === NodeType.MetricOp) {
      // Get the metrics id, and navigate to the metric details page.
      return (
        <Box display="flex" flexDirection="row">
          <Button
            style={buttonStyle}
            onClick={() => {
              navigate(
                `${getPathPrefix()}/workflow/${workflowId}/result/${
                  workflow.selectedResult.id
                }/metric/${currentNode.id}`
              );
            }}
          >
            View Metric Details
          </Button>
          {exportOpButton}
        </Box>
      );
    }

    if (currentNode.type === NodeType.FunctionOp) {
      return (
        <Box display="flex" flexDirection="row">
          <Button
            style={buttonStyle}
            onClick={() => {
              navigate(
                `${getPathPrefix()}/workflow/${workflowId}/result/${
                  workflow.selectedResult.id
                }/operator/${currentNode.id}`
              );
            }}
          >
            View Operator Details
          </Button>
          {exportOpButton}
        </Box>
      );
    }

    if (currentNode.type === NodeType.CheckOp) {
      return (
        <Box display="flex" flexDirection="row">
          <Button
            style={buttonStyle}
            onClick={() => {
              navigate(
                `${getPathPrefix()}/workflow/${workflowId}/result/${
                  workflow.selectedResult.id
                }/check/${currentNode.id}`
              );
            }}
          >
            View Check Details
          </Button>
          {exportOpButton}
        </Box>
      );
    }

    return null;
  };

  const drawerHeaderHeightInPx = 64;

  return (
    <Layout
      breadcrumbs={[
        BreadcrumbLink.HOME,
        BreadcrumbLink.WORKFLOWS,
        new BreadcrumbLink(path, dagName),
      ]}
      user={user}
      onBreadCrumbClicked={() => {
        resetWorkflowState();
      }}
      onSidebarItemClicked={() => {
        resetWorkflowState();
      }}
    >
      <Box
        sx={{
          boxSizing: 'border-box',
          display: 'flex',
          // TODO: just create a state variable to reflect the open state of the drawer.
          width:
            currentNode.type === NodeType.None
              ? "100%;"
              : `calc(100% - ${SidesheetWidth});`,
          height: '100%',
          flexDirection: 'column',
          transition: WidthTransition,
          transitionDelay: '-150ms',
        }}
        id={WorkflowPageContentId}
      >
        {workflow.selectedDag && (
          <Box marginBottom={1}>
            <WorkflowHeader
              user={user}
              workflowDag={workflow.selectedDag}
              workflowId={workflowId}
            />
          </Box>
        )}

        <Divider />

        <Box
          sx={{
            flex: 1,
            mt: 2,
            p: 3,
            mb: contentBottomOffsetInPx,
            width: '100%',
            boxSizing: 'border-box',
            backgroundColor: 'gray.50',
          }}
        >
          <ReactFlowProvider>
            <ReactFlowCanvas
              switchSideSheet={switchSideSheet}
              onPaneClicked={onPaneClicked}
            />
          </ReactFlowProvider>
        </Box>
      </Box>

      {currentNode.type !== NodeType.None && (
        <Drawer
          anchor="right"
          variant="persistent"
          open={true}
          PaperProps={{
            sx: {
              transition: 'width 200ms ease-in-out',
              transitionDelay: '1000ms',
            },
          }}
        >
          <Box
            width={SidesheetWidth}
            maxWidth={SidesheetWidth}
            minHeight="100vh"
            display="flex"
            flexDirection="column"
          >
            <Box
              width="100%"
              sx={{ backgroundColor: theme.palette.gray['100'] }}
              height={`${drawerHeaderHeightInPx}px`}
            >
              <Box display="flex">
                <Box
                  sx={{ cursor: 'pointer', m: 1, alignSelf: 'center' }}
                  onClick={onPaneClicked}
                >
                  <FontAwesomeIcon icon={faChevronRight} />
                </Box>
                <Box maxWidth="400px">
                  <Typography
                    variant="h5"
                    padding="16px"
                    textOverflow="ellipsis"
                    overflow="hidden"
                    whiteSpace="nowrap"
                  >
                    {getNodeLabel()}
                  </Typography>
                </Box>
                <Box sx={{ mx: 2, alignSelf: 'center' }}>
                  {getNodeActionButton()}
                </Box>
              </Box>
            </Box>
            <Box
              sx={{
                overflow: 'auto',
                flexGrow: 1,
                marginBottom: DefaultLayoutMargin,
              }}
            >
              {getDataSideSheetContent(
                user,
                currentNode,
                workflowId,
                workflow.selectedResult.id
              )}
            </Box>
          </Box>
        </Drawer>
      )}
    </Layout>
  );
};

export default WorkflowPage;
