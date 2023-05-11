import {
  faArrowRotateRight,
  faChevronRight,
  faCirclePlay,
  faUpRightAndDownLeftFromCenter,
} from '@fortawesome/free-solid-svg-icons';
import { faCircleDown } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Alert, Drawer, Snackbar, Tooltip } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { ReactFlowProvider } from 'reactflow';
import WorkflowResultNavigator from 'src/components/workflows/WorkflowResultNavigator';

import {
  aqueductApi,
  useDagResultGetQuery,
  useDagResultsGetQuery,
  useWorkflowGetQuery,
} from '../../../../handlers/AqueductApi';
import { handleLoadIntegrations } from '../../../../reducers/integrations';
import { NodeType } from '../../../../reducers/nodeSelection';
import { selectNode } from '../../../../reducers/pages/Workflow';
import { handleGetSelectDagPosition } from '../../../../reducers/workflow';
import { AppDispatch, RootState } from '../../../../stores/store';
import { theme } from '../../../../styles/theme/theme';
import UserProfile from '../../../../utils/auth';
import { handleExportFunction } from '../../../../utils/operators';
import { WidthTransition } from '../../../../utils/shared';
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
} from '../../../workflows/WorkflowHeader';
import WorkflowSettings from '../../../workflows/WorkflowSettings';
import { LayoutProps } from '../../types';
import RunWorkflowDialog from '../../workflows/components/RunWorkflowDialog';
import {
  useWorkflowBreadcrumbs,
  useWorkflowIds,
  useWorkflowNodes,
  useWorkflowNodesResults,
} from './hook';

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
  const { workflowId, dagId, dagResultId } = useWorkflowIds(user.apiKey);
  const breadcrumbs = useWorkflowBreadcrumbs(
    user.apiKey,
    workflowId,
    dagId,
    dagResultId,
    'Workflow'
  );
  const {
    data: workflow,
    isLoading: wfLoading,
    error: wfError,
  } = useWorkflowGetQuery(
    { apiKey: user.apiKey, workflowId },
    { skip: !workflowId }
  );
  const { data: dagResult } = useDagResultGetQuery(
    { apiKey: user.apiKey, workflowId, dagResultId },
    { skip: !workflowId || !dagResultId }
  );
  const { data: dagResults } = useDagResultsGetQuery(
    { apiKey: user.apiKey, workflowId },
    { skip: !workflowId }
  );
  const nodes = useWorkflowNodes(user.apiKey, workflowId, dagId);
  const nodeResults = useWorkflowNodesResults(
    user.apiKey,
    workflowId,
    dagResultId
  );

  const [currentTab, setCurrentTab] = useState<string>('Details');
  const [showRunWorkflowDialog, setShowRunWorkflowDialog] = useState(false);

  const [updateMessage, setUpdateMessage] = useState<string>('');
  const [showUpdateMessage, setShowUpdateMessage] = useState<boolean>(false);
  const [updateSucceeded, setUpdateSucceeded] = useState<boolean>(false);

  const selectedNodeState = useSelector(
    (state: RootState) =>
      state.workflowPageReducer.perWorkflowPageStates[workflowId]?.SelectedNode
  );

  const selectedNode =
    nodes[selectedNodeState.nodeType][selectedNodeState.nodeId];
  const selectedNodeResult =
    nodeResults[selectedNodeState.nodeType][selectedNodeState.nodeId];

  const switchSideSheet = sideSheetSwitcher(dispatch);
  const drawerIsOpen = !!selectedNode;

  const onPaneClicked = (event: React.MouseEvent) => {
    event.preventDefault();

    // Reset selected node
    dispatch(selectNode({ workflowId, selection: undefined }));
  };

  useEffect(() => {
    if (workflow !== undefined) {
      document.title = `${workflow.name} | Aqueduct`;
    }
  }, [workflow]);

  // Load Integrations
  useEffect(() => {
    dispatch(handleLoadIntegrations({ apiKey: user.apiKey }));
  }, [dispatch, user.apiKey, workflowId]);

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

  // This workflow doesn't exist.
  if (wfError) {
    navigate('/404');
    return null;
  }

  if (wfLoading) {
    return null;
  }

  const nodeLabel =
    selectedNode.name ??
    (selectedNodeState.nodeType === 'operators'
      ? 'Operator Node'
      : 'Artifact Node');
  const getNodeActionButton = () => {
    const buttonStyle = {
      fontSize: '20px',
      mr: 1,
    };

    let navigateButton;
    let includeExportOpButton = true;

    if (!dagResultId) {
      return null;
    } else {
      let navigationUrl;
      if (currentNode.type === NodeType.TableArtifact) {
        navigationUrl = `/workflow/${workflowId}/result/${dagResultId}/artifact/${currentNode.id}`;
        includeExportOpButton = false;
      } else if (currentNode.type === NodeType.FunctionOp) {
        navigationUrl = `/workflow/${workflowId}/result/${dagResultId}/operator/${currentNode.id}`;
      } else if (currentNode.type === NodeType.MetricOp) {
        navigationUrl = `/workflow/${workflowId}/result/${dagResultId}/metric/${currentNode.id}`;
      } else if (currentNode.type === NodeType.CheckOp) {
        navigationUrl = `/workflow/${workflowId}/result/${dagResultId}/check/${currentNode.id}`;
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
      breadcrumbs={breadcrumbs}
      user={user}
      onBreadCrumbClicked={() => {
        return;
      }}
      onSidebarItemClicked={() => {
        return;
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
        <Box marginBottom={1}>
          <WorkflowHeader
            apiKey={user.apiKey}
            workflowId={workflowId}
            dagId={dagId}
            dagResultId={dagResultId}
          />
        </Box>

        {/*Show any workflow-level errors at the top of the workflow details page.*/}
        {dagResult?.exec_state?.error && (
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
            >{`${dagResult.exec_state.error.tip}\n\n${dagResult.exec_state.error.context}`}</pre>
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
            {dagResults !== undefined && dagResults.length > 0 && (
              <Box
                mb={2}
                pb={2}
                width="100%"
                sx={{ borderBottom: `1px solid ${theme.palette.gray[600]}` }}
              >
                <WorkflowResultNavigator apiKey={user.apiKey} />
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
                  // refresh node results, result history, and current result
                  dispatch(
                    aqueductApi.endpoints.nodesResultsGet.initiate({
                      apiKey: user.apiKey,
                      workflowId,
                      dagResultId,
                    })
                  );

                  dispatch(
                    aqueductApi.endpoints.dagResultGet.initiate({
                      apiKey: user.apiKey,
                      workflowId,
                      dagResultId,
                    })
                  );

                  dispatch(
                    aqueductApi.endpoints.dagResultsGet.initiate({
                      apiKey: user.apiKey,
                      workflowId,
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
                  {nodeLabel}
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
              dagId,
              dagResultId
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
