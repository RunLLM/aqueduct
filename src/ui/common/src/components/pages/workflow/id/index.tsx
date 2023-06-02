import {
  faArrowRotateRight,
  faChevronRight,
  faCirclePlay,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Alert, Drawer, Snackbar, Tooltip } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { ReactFlowProvider } from 'reactflow';

import WorkflowResultNavigator from '../../../../components/workflows/WorkflowResultNavigator';
import {
  useDagGetQuery,
  useDagResultGetQuery,
  useDagResultsGetQuery,
  useNodesResultsGetQuery,
  useWorkflowEditPostMutation,
  useWorkflowGetQuery,
} from '../../../../handlers/AqueductApi';
import { selectNode } from '../../../../reducers/pages/Workflow';
import { handleLoadResources } from '../../../../reducers/resources';
import { AppDispatch, RootState } from '../../../../stores/store';
import { theme } from '../../../../styles/theme/theme';
import UserProfile from '../../../../utils/auth';
import { WidthTransition } from '../../../../utils/shared';
import { getDataSideSheetContent } from '../../../../utils/sidesheets';
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
import WorkflowNodeSidesheetActions from '../../../workflows/WorkflowNodeSidesheetActions';
import WorkflowSettings from '../../../workflows/WorkflowSettings';
import { LayoutProps } from '../../types';
import RunWorkflowDialog from '../../workflows/components/RunWorkflowDialog';
import {
  useSortedDagResults,
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
    refetch: refetchWorkflow,
  } = useWorkflowGetQuery(
    { apiKey: user.apiKey, workflowId },
    { skip: !workflowId }
  );
  const { data: dag, isLoading: dagLoading } = useDagGetQuery(
    { apiKey: user.apiKey, workflowId, dagId },
    { skip: !workflowId || !dagId }
  );
  const { data: dagResult, refetch: refetchDagResult } = useDagResultGetQuery(
    { apiKey: user.apiKey, workflowId, dagResultId },
    { skip: !workflowId || !dagResultId }
  );
  const dagResults = useSortedDagResults(user.apiKey, workflowId);
  const { refetch: refetchDagResults } = useDagResultsGetQuery(
    { apiKey: user.apiKey, workflowId },
    { skip: !workflowId }
  );
  const { refetch: refetchNodeResults } = useNodesResultsGetQuery(
    { apiKey: user.apiKey, workflowId, dagResultId },
    { skip: !workflowId || !dagResultId }
  );

  const nodes = useWorkflowNodes(user.apiKey, workflowId, dagId);
  const nodeResults = useWorkflowNodesResults(
    user.apiKey,
    workflowId,
    dagResultId
  );

  const [currentTab, setCurrentTab] = useState<string>('Details');
  const [showRunWorkflowDialog, setShowRunWorkflowDialog] = useState(false);

  const [
    {},
    {
      isSuccess: editWorkflowSuccess,
      error: editWorkflowError,
      reset: resetEditWorkflow,
    },
  ] = useWorkflowEditPostMutation({
    fixedCacheKey: `edit-${workflowId}`,
  });

  const editWorkflowMessage = editWorkflowSuccess
    ? 'Sucessfully updated your workflow.'
    : editWorkflowError
    ? `There was an unexpected error while updating your workflow: ${editWorkflowError}`
    : '';

  const selectedNodeState = useSelector(
    (state: RootState) =>
      state.workflowPageReducer.perWorkflowPageStates[workflowId]?.SelectedNode
  );

  const selectedNode = !!selectedNodeState
    ? nodes[selectedNodeState.nodeType][selectedNodeState.nodeId]
    : undefined;

  const drawerIsOpen = !!selectedNode;

  useEffect(() => {
    if (workflow !== undefined) {
      document.title = `${workflow.name} | Aqueduct`;
    }
  }, [workflow]);

  // Load Resources
  useEffect(() => {
    dispatch(handleLoadResources({ apiKey: user.apiKey }));
  }, [dispatch, user.apiKey, workflowId]);

  useEffect(() => {
    if (editWorkflowSuccess) {
      refetchWorkflow();
      setCurrentTab('Details');
    }
  }, [refetchWorkflow, editWorkflowSuccess]);

  // This workflow doesn't exist.
  if (wfError) {
    navigate('/404');
    return null;
  }

  if (wfLoading || dagLoading) {
    return null;
  }

  const nodeLabel =
    selectedNode?.name ??
    (selectedNodeState?.nodeType === 'operators'
      ? 'Operator Node'
      : 'Artifact Node');

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
            {currentTab === 'Details' && !!dag && !!nodes && (
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
                      nodes={nodes}
                      nodeResults={nodeResults}
                      dag={dag}
                    />
                  </Box>
                </ReactFlowProvider>
              </Box>
            )}

            {currentTab === 'Settings' && !!workflow && !!dag && (
              <Box sx={{ paddingBottom: '24px' }}>
                <WorkflowSettings user={user} dag={dag} workflow={workflow} />
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
                  refetchDagResult();
                  refetchDagResults();
                  refetchNodeResults();
                }}
              >
                <FontAwesomeIcon icon={faArrowRotateRight} />
              </Button>
            </Tooltip>
          </Box>
        </Box>

        {!!nodes && !!workflow && (
          <RunWorkflowDialog
            user={user}
            nodes={nodes}
            workflowId={workflowId}
            open={showRunWorkflowDialog}
            setOpen={setShowRunWorkflowDialog}
            name={workflow.name}
          />
        )}
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
            <Box display="flex" mr={3}>
              <Box
                sx={{ cursor: 'pointer', m: 1, alignSelf: 'center' }}
                onClick={() =>
                  dispatch(selectNode({ workflowId, selection: undefined }))
                }
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
              {/* This flex grown box right aligns the buttons below.*/}
              <Box flex={1} />

              {dagResultId && !!selectedNode && !!selectedNodeState && (
                <WorkflowNodeSidesheetActions
                  user={user}
                  workflowId={workflowId}
                  dagResultId={dagResultId}
                  selectedNodeState={selectedNodeState}
                  selectedNode={selectedNode}
                />
              )}
            </Box>
          </Box>
          {selectedNodeState && selectedNode && (
            <Box
              sx={{
                overflow: 'auto',
                flexGrow: 1,
                marginBottom: DefaultLayoutMargin,
              }}
            >
              {getDataSideSheetContent(user, selectedNodeState, selectedNode)}
            </Box>
          )}
        </Box>
      </Drawer>

      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={!!editWorkflowMessage}
        onClose={() => resetEditWorkflow()}
        key={'settingsupdate-snackbar'}
        autoHideDuration={3000}
      >
        <Alert
          onClose={() => resetEditWorkflow()}
          severity={editWorkflowSuccess ? 'success' : 'error'}
          sx={{ width: '100%' }}
        >
          {editWorkflowMessage}
        </Alert>
      </Snackbar>
    </Layout>
  );
};

export default WorkflowPage;
