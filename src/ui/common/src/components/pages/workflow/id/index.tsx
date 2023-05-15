import {
  faArrowRotateRight,
  faChevronLeft,
  faChevronRight,
  faCirclePlay,
  faUpRightAndDownLeftFromCenter,
} from '@fortawesome/free-solid-svg-icons';
import { faCircleDown } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Alert, Drawer, Snackbar, Tooltip } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import { parse } from 'query-string';
import React, { useCallback, useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useLocation, useNavigate, useParams } from 'react-router-dom';
import { ReactFlowProvider } from 'reactflow';

import { BreadcrumbLink } from '../../../../components/layouts/NavBar';
import { handleGetWorkflowHistory } from '../../../../handlers/getWorkflowHistory';
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
import { Tab, Tabs } from '../../../primitives/Tabs.styles';
import ReactFlowCanvas from '../../../workflows/ReactFlowCanvas';
import WorkflowHeader, {
  WorkflowPageContentId,
} from '../../../workflows/workflowHeader';
import WorkflowSettings from '../../../workflows/WorkflowSettings';
import { LayoutProps } from '../../types';
import RunWorkflowDialog from '../../workflows/components/RunWorkflowDialog';
import { useNodeArtifactGetQuery, useNodeOperatorGetQuery, useNodesGetQuery, useNodesResultsGetQuery } from '../../../../handlers/AqueductApi';


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

  const [updateMessage, setUpdateMessage] = useState<string>('');
  const [showUpdateMessage, setShowUpdateMessage] = useState<boolean>(false);
  const [updateSucceeded, setUpdateSucceeded] = useState<boolean>(false);

  const currentNode = useSelector(
    (state: RootState) => state.nodeSelectionReducer.selected
  );
  const workflow = useSelector((state: RootState) => state.workflowReducer);
  const switchSideSheet = sideSheetSwitcher(dispatch);
  const drawerIsOpen = currentNode.type !== NodeType.None;

  const dagName = workflow.selectedDag?.metadata?.name;


  // const { data: NodesData, error: NodesError, isLoading: NodesLoading } = useNodesGetQuery(
  //   {
  //     apiKey: user.apiKey,
  //     workflowId: workflowId,
  //     dagId: workflow.selectedResult?.id,
  //   },
  //   { pollingInterval: 5000 }
  // )
  // console.log("NodesData", NodesData);

  // const { data: NodesResultData, error: NodesResultError, isLoading: NodesResultLoading } = useNodesResultsGetQuery(
  //   {
  //     apiKey: user.apiKey,
  //     workflowId: workflowId,
  //     dagId: workflow.selectedResult?.id,
  //   },
  //   { pollingInterval: 5000 }
  // )
  // console.log("NodesResultData", NodesResultData);
  
  // const { data: NodesArtifactData, error: NodesArtifactError, isLoading: NodesArtifactLoading } = useNodeArtifactGetQuery(
  //   {
  //     apiKey: user.apiKey,
  //     workflowId: workflowId,
  //     dagId: workflow.selectedResult?.id,
  //     nodeId: Object.keys(workflow.selectedDag?.artifacts)[0],
  //   },
  //   { pollingInterval: 5000 }
  // )
  // console.log("NodesArtifactData", NodesArtifactData);
  
  // const { data: NodesOpData, error: NodesOpError, isLoading: NodesOpLoading } = useNodeOperatorGetQuery(
  //   {
  //     apiKey: user.apiKey,
  //     workflowId: workflowId,
  //     dagId: workflow.selectedResult.id,
  //     nodeId: Object.keys(workflow.selectedDag?.operators)[0],
  //   },
  //   { pollingInterval: 5000 }
  // )
  // console.log("NodesOpData", NodesOpData);
  

  // TODO: Add metrics and checks to useNodesGetQuery & useNodesResultsGetQuery
  // TODO: Remove metrics & checks from useNodeArtifactGetQuery & useNodeOperatorGetQuery & related queries
  // TODO: Create useNodeMetricGetQuery & useNodeCheckGetQuery + the contents



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

      if (
        !!workflow.selectedResult &&
        workflowDagResultId !== workflow.selectedResult.id
      ) {
        // this is where selectedDag gets set
        dispatch(selectResultIdx(workflowDagResultIndex));
      }

      // This is outside the if statement because this is not automatically kept in sync with
      // the Redux store.
      setSelectedResultIdx(workflowDagResultIndex);
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

  useEffect(() => {
    dispatch(
      handleGetWorkflowHistory({
        apiKey: user.apiKey,
        workflowId: workflowId,
      })
    );
  }, [user.apiKey]);

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

    let navigateButton;
    let includeExportOpButton = true;

    if (!workflow.selectedResult) {
      return null;
    } else {
      let navigationUrl;
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

      navigateButton = (
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
    }

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
        {/* This flex grown box right aligns the two buttons below.*/}
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

        {/*Show any workflow-level errors at the top of the workflow details page.*/}
        {workflow.selectedResult?.exec_state?.error && (
          <Box
            sx={{
              backgroundColor: theme.palette.red[100],
              color: theme.palette.red[600],
              p: 2,
              paddingBottom: '16px',
              paddingTop: '16px',
              height: 'fit-content',

              // When the sidesheet is not open, we want to align the right side with the
              // dag viewer. This means taking off 100px (the width of the right control column)
              // + 16px the left margin of the control column
              // + 32px the additional width to the end of the screen.
              // When the sidesheet is open, the control plane disappears, so we just need
              // the last adjustment of 32px.
              width: !drawerIsOpen ? `calc(100% - 148px)` : 'calc(100% - 32px)',
            }}
          >
            <pre
              style={{ margin: '0px' }}
            >{`${workflow.selectedResult.exec_state.error.tip}\n\n${workflow.selectedResult.exec_state.error.context}`}</pre>
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
                  onSettingsSave={() => {
                    setShowUpdateMessage(true);
                    // Show toast message for a few seconds and then update the current tab.
                    setTimeout(() => {
                      // Refresh the page to send user to Details tab with latest information.
                      window.location.reload();
                    }, 3000);
                  }}
                  onSetShowUpdateMessage={(shouldShow) =>
                    setShowUpdateMessage(shouldShow)
                  }
                  onSetUpdateSucceeded={(isSuccessful) =>
                    setUpdateSucceeded(isSuccessful)
                  }
                  onSetUpdateMessage={(updateMessage) =>
                    setUpdateMessage(updateMessage)
                  }
                />
              </Box>
            )}
          </Box>

          {/* These controls are automatically hidden when the side sheet is open. */}
          {/* Tooltips don't show up if the child is disabled so we wrap the button with a Box.  */}
          <Box width="100px" ml={2} display={drawerIsOpen ? 'none' : 'block'}>
            {workflow.dagResults && workflow.dagResults.length > 1 && (
              <Box
                display="flex"
                mb={2}
                pb={2}
                width="100%"
                sx={{ borderBottom: `1px solid ${theme.palette.gray[600]}` }}
              >
                <Tooltip title="Previous Run" arrow>
                  <Box sx={{ px: 0, flex: 1 }}>
                    <Button
                      sx={{ fontSize: '28px', width: '100%' }}
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
                  </Box>
                </Tooltip>

                <Tooltip title="Next Run" arrow>
                  <Box sx={{ px: 0, flex: 1 }}>
                    <Button
                      sx={{ fontSize: '28px', width: '100%' }}
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
                  </Box>
                </Tooltip>
              </Box>
            )}

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
                  // When the button is clicked, load all of the metadata of each of the nodes again.
                  getDagResultDetails(true);

                  // Also refresh the history of workflow runs to update the status.
                  dispatch(
                    handleGetWorkflowHistory({
                      apiKey: user.apiKey,
                      workflowId: workflowId,
                    })
                  );
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
            sx={{ backgroundColor: theme.palette.gray[100] }}
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
              selectedDag.id,
              workflow.selectedResult?.id
            )}
          </Box>
        </Box>
      </Drawer>

      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showUpdateMessage}
        onClose={() => setShowUpdateMessage(false)}
        key={'settingsupdate-snackbar'}
        autoHideDuration={3000}
      >
        <Alert
          onClose={() => setShowUpdateMessage(false)}
          severity={updateSucceeded ? 'success' : 'error'}
          sx={{ width: '100%' }}
        >
          {updateMessage}
        </Alert>
      </Snackbar>
    </Layout>
  );
};

export default WorkflowPage;
