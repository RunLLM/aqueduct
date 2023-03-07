import {
  faArrowRotateRight,
  faChevronLeft,
  faChevronRight,
  faCirclePlay,
  faUpRightAndDownLeftFromCenter,
} from '@fortawesome/free-solid-svg-icons';
import { faCircleDown } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Drawer, Tooltip } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import { parse } from 'query-string';
import React, { useCallback, useEffect, useState } from 'react';
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
import { handleExportFunction } from '../../../../utils/operators';
import { LoadingStatusEnum, WidthTransition } from '../../../../utils/shared';
import {
  getDataSideSheetContent,
  sideSheetSwitcher,
} from '../../../../utils/sidesheets';
import DefaultLayout, {
  DefaultLayoutMargin,
  SidesheetWidth,
} from '../../../layouts/default';
import { Button } from '../../../primitives/Button.styles';
import { Tab, Tabs } from '../../../Tabs/Tabs.styles';
import ReactFlowCanvas from '../../../workflows/ReactFlowCanvas';
import WorkflowHeader, {
  WorkflowPageContentId,
} from '../../../workflows/workflowHeader';
import WorkflowSettings from '../../../workflows/WorkflowSettings';
import { LayoutProps } from '../../types';
import RunWorkflowDialog from '../../workflows/components/RunWorkflowDialog';

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

  const [currentTab, setCurrentTab] = useState<string>('Details');
  const [showRunWorkflowDialog, setShowRunWorkflowDialog] = useState(false);
  const [selectedResultIdx, setSelectedResultIdx] = useState(0);

  const currentNode = useSelector(
    (state: RootState) => state.nodeSelectionReducer.selected
  );
  const workflow = useSelector((state: RootState) => state.workflowReducer);
  const switchSideSheet = sideSheetSwitcher(dispatch);
  const drawerIsOpen = currentNode.type !== NodeType.None;

  const dagName = workflow.selectedDag?.metadata?.name;

  // EFFECT 0: Set document title.
  useEffect(() => {
    if (workflow.selectedDag !== undefined) {
      document.title = `${dagName} | Aqueduct`;
    }
  }, [workflow.selectedDag, dagName]);

  const resetWorkflowState = useCallback(() => {
    dispatch(resetState());
  }, [dispatch]);

  // EFFECT 1: Manage state on browser history change.
  // This effect adds the resetWorkflowState callback to be used when the user
  // accesses the page history. In this case, we reset the state of the Redux
  // workflow store, so we don't accidentally cache information across workflow
  // versions.
  useEffect(() => {
    window.onpopstate = () => {
      resetWorkflowState();
    };

    resetWorkflowState();
  }, [resetWorkflowState]);

  // EFFECT 2: Set URL search param on version change.
  // When the selected workflow run changes, we update the URL search param accordingly.
  // This is important for two reasons:
  // 1. It makes the URL sharable.
  // 2. We rely on this to track what workflow version we're currently displaying.
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

  // EFFECT 3: Load workflow metadata.
  // This useEffect is effectively only called on component mount. It loads
  // the base workflow metadata as well as metadata about any integrations
  // in order to populate the UI.
  useEffect(() => {
    dispatch(handleGetWorkflow({ apiKey: user.apiKey, workflowId }));
    dispatch(handleLoadIntegrations({ apiKey: user.apiKey }));
  }, [dispatch, user.apiKey, workflowId]);

  // EFFECT 4: Gather selected workflow index.
  // When the workflow Redux store's DAG results are populated or when we navigate
  // to a different version, we iterate through the full list of results and set the
  // index in both Redux and in our local state.
  // NOTE(vikram): There are two annoying bits of tech debt in this code:
  // 1. It's not clear that this needs to be a different from Effect 2 the one where we
  // navigate to a different search param. They seem to be focused on the same bits of
  // functionality. (See ENG-2569.)
  // 2. Less critical, but it's annoying that we have to track selectedResultIdx in local
  // React state. This is not explicitly exposed by the Redux store, but it should be.
  useEffect(() => {
    if (workflow.dagResults && workflow.dagResults.length > 0) {
      let workflowDagResultIndex = 0;
      const { workflowDagResultId } = urlSearchParams;

      // Iterate through all the results and check which one's ID matches the ID of
      // the Redux store's selected DAG result.
      for (let i = 0; i < workflow.dagResults.length; i++) {
        if (workflow.dagResults[i].id === workflowDagResultId) {
          workflowDagResultIndex = i;
          break;
        }
      }

      if (workflowDagResultId !== workflow.selectedResult.id) {
        // this is where selectedDag gets set
        dispatch(selectResultIdx(workflowDagResultIndex));
        setSelectedResultIdx(workflowDagResultIndex);
      }
    }
  }, [
    workflow.dagResults,
    urlSearchParams,
    workflow.selectedResult?.id,
    dispatch,
  ]);

  // EFFECT 5: DAG positioning.
  // This effect uses the Elk algorithm to load the node positioning for the DAG.
  // See ENG-2568 for more on how this interaction needs to be cleaned up.
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
   * @param metadataOnly if set to to true, only the status of the artifact
   * will be retrieved but the data itself will be skipped
   * @param force whether to reload the results regardless of whether
   * they are cached.
   */
  const getArtifactResultDetails = useCallback(
    (nodeId: string, metadataOnly: boolean, force = false) => {
      const artf = (workflow.selectedDag?.artifacts ?? {})[nodeId];
      if (!artf || !workflow.selectedResult) {
        return;
      }

      if (!(nodeId in workflow.artifactResults) || force) {
        dispatch(
          handleGetArtifactResults({
            apiKey: user.apiKey,
            workflowDagResultId: workflow.selectedResult.id,
            artifactId: nodeId,
            metadataOnly: metadataOnly,
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
   * @param force whether to reload the results regardless of whether
   * they are cached.
   */
  const getOperatorResultDetails = useCallback(
    (nodeId: string, force = false) => {
      // Verify the node is indeed an operator, and a result is selected
      const op = (workflow.selectedDag?.operators ?? {})[nodeId];
      if (!op || !workflow.selectedResult) {
        return;
      }

      if (!(nodeId in workflow.operatorResults) || force) {
        dispatch(
          handleGetOperatorResults({
            apiKey: user.apiKey,
            workflowDagResultId: workflow.selectedResult.id,
            operatorId: nodeId,
          })
        );
      }

      if (op.spec.metric || op.spec.check) {
        for (const artfId of [...op.outputs]) {
          // We set metadataOnly to false because for metric and check, we want to also show
          // their values on the workflow page.
          getArtifactResultDetails(artfId, false, force);
        }
      } else {
        for (const artfId of [...op.outputs]) {
          getArtifactResultDetails(artfId, true, force);
        }
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

  // EFFECT 6: Load operator and artifact metadta.
  // This effect loads the relevant metadata for each operator and artifact when either the
  // selected node changes or when the selected workflow run changes. This is probably a
  // little sloppy at the moment because we're pushing error checks into the helper functions
  // and blindly calling them, which is opaque/confusing to read.
  useEffect(() => {
    getOperatorResultDetails(currentNode.id);
    getArtifactResultDetails(currentNode.id, true);
  }, [
    currentNode?.id,
    getArtifactResultDetails,
    getOperatorResultDetails,
    workflow.selectedResult?.id,
  ]);

  const onPaneClicked = (event: React.MouseEvent) => {
    event.preventDefault();

    // Reset selected node
    dispatch(resetSelectedNode());
  };

  const selectedDag = workflow.selectedDag;
  // This function retrieves all of the node metadata in this workflow DAG result.
  // The `force` flag forces a reload even if the data is already present. This is
  // used to refresh the state of a DAG that's already been loaded.
  const getDagResultDetails = (force = false) => {
    if (
      (workflow.loadingStatus.loading === LoadingStatusEnum.Succeeded &&
        !!selectedDag) ||
      force
    ) {
      for (const op of Object.values(selectedDag.operators)) {
        // We don't need to call getArtifactResultDetails because
        // getOperatorResultDetails automatically does that for us.
        getOperatorResultDetails(op.id, force);
      }
    }
  };

  // EFFECT 7: Load full DAG metadata.
  // This effect loads all of the metadata associated with a particular workflow run.
  // Both this an Effect 6 are run on every workflow run ID change, which might be
  // duplicative (ENG-2569).
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
      fontSize: '20px',
      mr: 1,
    };

    let navigationUrl;
    let includeExportOpButton = true;

    if (currentNode.type === NodeType.TableArtifact) {
      navigationUrl = `/workflow/${workflowId}/result/${workflow.selectedResult.id}/artifact/${currentNode.id}`;
      includeExportOpButton = false;
    } else if (currentNode.type === NodeType.FunctionOp) {
      navigationUrl = `/workflow/${workflowId}/result/${workflow.selectedResult.id}/operator/${currentNode.id}`;
    } else if (currentNode.type === NodeType.MetricOp) {
      navigationUrl = `/workflow/${workflowId}/result/${workflow.selectedResult.id}/metric/${currentNode.id}`;
    } else if (currentNode.type === NodeType.CheckOp) {
      navigationUrl = `/workflow/${workflowId}/result/${workflow.selectedResult.id}/check/${currentNode.id}`;
    } else {
      return null; // This is a load or save operator.
    }

    const navigateButton = (
      <Button
        variant="text"
        sx={buttonStyle}
        onClick={() => {
          navigate(navigationUrl);
        }}
      >
        <Tooltip title="Expand Details" arrow>
          <FontAwesomeIcon icon={faUpRightAndDownLeftFromCenter} />
        </Tooltip>
      </Button>
    );

    const operator = (workflow.selectedDag?.operators ?? {})[currentNode.id];
    const exportOpButton = (
      <Button
        onClick={async () => {
          await handleExportFunction(
            user,
            currentNode.id,
            `${operator?.name ?? 'function'}.zip`
          );
        }}
        variant="text"
        sx={buttonStyle}
      >
        <Tooltip title="Download Code" arrow>
          <FontAwesomeIcon icon={faCircleDown} />
        </Tooltip>
      </Button>
    );

    return (
      <Box display="flex" alignItems="center" flex={1} mr={3}>
        {/* This flex grown box right aligns the bwo buttons below.*/}
        <Box flex={1} />
        <Box display="flex" alignItems="center">
          {includeExportOpButton && exportOpButton}
          {navigateButton}
        </Box>
      </Box>
    );
  };

  const drawerHeaderHeightInPx = 64;

  const handleTabChange = (event: React.SyntheticEvent, newTab: string) => {
    setCurrentTab(newTab);
  };

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
          width: !drawerIsOpen ? '100%;' : `calc(100% - ${SidesheetWidth});`,
          height: '100%',
          flexDirection: 'column',
          transition: WidthTransition,
          transitionDelay: '-150ms',
          paddingBottom: '24px',
        }}
        id={WorkflowPageContentId}
      >
        {workflow.selectedDag && (
          <Box marginBottom={1}>
            <WorkflowHeader workflowDag={workflow.selectedDag} />
          </Box>
        )}

        <Tabs value={currentTab} onChange={handleTabChange} sx={{ mb: 1 }}>
          <Tab value="Details" label="Details" />
          <Tab value="Settings" label="Settings" />
        </Tabs>

        <Box display="flex" height="100%">
          <Box flex={1} height="100%">
            {currentTab === 'Details' && (
              <Box
                sx={{
                  flexDirection: 'column',
                  display: 'flex',
                  flexGrow: 1,
                  height: '100%',
                  backgroundColor: theme.palette.gray[50],
                }}
              >
                <ReactFlowProvider>
                  <Box sx={{ flexGrow: 1 }}>
                    <ReactFlowCanvas
                      switchSideSheet={switchSideSheet}
                      onPaneClicked={onPaneClicked}
                    />
                  </Box>
                </ReactFlowProvider>
              </Box>
            )}

            {currentTab === 'Settings' && workflow.selectedDag && (
              <Box sx={{ paddingBottom: '24px' }}>
                <WorkflowSettings
                  user={user}
                  workflowDag={workflow.selectedDag}
                />
              </Box>
            )}
          </Box>

          {/* These controls are automatically hidden when the side sheet is open. */}
          <Box width="100px" ml={2} display={drawerIsOpen ? 'none' : 'block'}>
            <Box
              display="flex"
              mb={2}
              pb={2}
              width="100%"
              sx={{ borderBottom: `1px solid ${theme.palette.gray[600]}` }}
            >
              <Tooltip title="Previous Run" arrow>
                <Button
                  sx={{ fontSize: '28px', px: 0, flex: 1 }}
                  variant="text"
                  onClick={() => {
                    // This might be confusing, but index 0 is the most recent run, so incrementing the index goes
                    // to an *earlier* run.
                    dispatch(selectResultIdx(selectedResultIdx + 1));
                    navigate(
                      `?workflowDagResultId=${
                        workflow.dagResults[selectedResultIdx + 1].id
                      }`
                    );
                  }}
                  disabled={
                    selectedResultIdx === workflow.dagResults.length - 1
                  }
                >
                  <FontAwesomeIcon icon={faChevronLeft} />
                </Button>
              </Tooltip>

              <Tooltip title="Next Run" arrow>
                <Button
                  sx={{ fontSize: '28px', px: 0, flex: 1 }}
                  variant="text"
                  onClick={() => {
                    // This might be confusing, but index 0 is the most recent run, so decrementing the index goes
                    // to a *newer* run.
                    dispatch(selectResultIdx(selectedResultIdx - 1));
                    navigate(
                      `?workflowDagResultId=${
                        workflow.dagResults[selectedResultIdx - 1].id
                      }`
                    );
                  }}
                  disabled={selectedResultIdx === 0}
                >
                  <FontAwesomeIcon icon={faChevronRight} />
                </Button>
              </Tooltip>
            </Box>

            <Box
              mb={2}
              pb={2}
              width="100%"
              sx={{ borderBottom: `1px solid ${theme.palette.gray[600]}` }}
            >
              <Tooltip title="Run Workflow" arrow>
                <Button
                  sx={{ width: '100%', py: 1, fontSize: '32px' }}
                  variant="text"
                  onClick={() => setShowRunWorkflowDialog(true)}
                >
                  <FontAwesomeIcon icon={faCirclePlay} />
                </Button>
              </Tooltip>
            </Box>

            <Tooltip title="Refresh" arrow>
              <Button
                sx={{ width: '100%', py: 1, fontSize: '32px' }}
                variant="text"
                onClick={() => {
                  // When the button is clicked, load all of the metadata again.
                  getDagResultDetails(true);
                }}
              >
                <FontAwesomeIcon icon={faArrowRotateRight} />
              </Button>
            </Tooltip>
          </Box>
        </Box>

        <RunWorkflowDialog
          user={user}
          workflowDag={workflow.selectedDag}
          workflowId={workflowId}
          open={showRunWorkflowDialog}
          setOpen={setShowRunWorkflowDialog}
        />
      </Box>

      <Drawer
        anchor="right"
        variant="persistent"
        open={drawerIsOpen}
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

              {getNodeActionButton()}
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
    </Layout>
  );
};

export default WorkflowPage;
