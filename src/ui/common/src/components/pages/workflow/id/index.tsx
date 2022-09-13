import { faChevronRight } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Drawer } from '@mui/material';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import Typography from '@mui/material/Typography';
import { parse } from 'query-string';
import React, { useEffect } from 'react';
import { ReactFlowProvider } from 'react-flow-renderer';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate, useParams } from 'react-router-dom';

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
  selectResultIdx,
} from '../../../../reducers/workflow';
import { AppDispatch, RootState } from '../../../../stores/store';
import { theme } from '../../../../styles/theme/theme';
import UserProfile from '../../../../utils/auth';
import { Data } from '../../../../utils/data';
import { exportCsv } from '../../../../utils/preview';
import { LoadingStatusEnum } from '../../../../utils/shared';
import { ExecutionStatus } from '../../../../utils/shared';
import {
  getDataSideSheetContent,
  sideSheetSwitcher,
} from '../../../../utils/sidesheets';
import DefaultLayout from '../../../layouts/default';
import { Button } from '../../../primitives/Button.styles';
import ReactFlowCanvas from '../../../workflows/ReactFlowCanvas';
import WorkflowStatusBar from '../../../workflows/StatusBar';
import WorkflowHeader from '../../../workflows/workflowHeader';
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

  const currentNode = useSelector(
    (state: RootState) => state.nodeSelectionReducer.selected
  );
  const workflow = useSelector((state: RootState) => state.workflowReducer);

  const switchSideSheet = sideSheetSwitcher(dispatch);
  const openSideSheetState = useSelector(
    (state: RootState) => state.openSideSheetReducer
  );
  const artifactResult = useSelector(
    (state: RootState) => state.workflowReducer.artifactResults[currentNode.id]
  );
  const dagPosition = useSelector(
    (state: RootState) => state.workflowReducer.selectedDagPosition
  );

  useEffect(() => {
    if (workflow.selectedDag !== undefined) {
      document.title = `${workflow.selectedDag.metadata.name} | Aqueduct`;
    }
  }, [workflow.selectedDag]);

  useEffect(() => {
    const urlSearchParams = parse(window.location.search);
    if (
      workflow.selectedResult !== undefined &&
      !urlSearchParams.workflowDagResultId
    ) {
      navigate(`?workflowDagResultId=${encodeURI(workflow.selectedResult.id)}`);
    }
  }, [workflow.selectedResult]);

  useEffect(() => {
    dispatch(handleGetWorkflow({ apiKey: user.apiKey, workflowId }));
    dispatch(handleLoadIntegrations({ apiKey: user.apiKey }));
  }, []);

  useEffect(() => {
    if (workflow.dagResults && workflow.dagResults.length > 0) {
      let workflowDagResultIndex = 0;
      const { workflowDagResultId } = parse(window.location.search);
      for (let i = 0; i < workflow.dagResults.length; i++) {
        if (workflow.dagResults[i].id === workflowDagResultId) {
          workflowDagResultIndex = i;
        }
      }
      if (workflowDagResultId !== workflow.selectedResult.id) {
        dispatch(setAllSideSheetState(false));
        dispatch(selectResultIdx(workflowDagResultIndex));
      }
    }
  }, [workflow.dagResults, window.location.search]);

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
  }, [workflow.selectedDag]);

  /**
   * This function dispatches calls to fetch artifact results and contents.
   *
   * This function is only activated when another similar fetch request
   * hasn't already been triggered.
   *
   * @param nodeId the UUID of the artifact for which we're retrieving
   * details.
   */
  const updateArtifactDetails = (nodeId: string) => {
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
  };

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
  const updateOperatorDetails = (nodeId: string) => {
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
      updateArtifactDetails(artfId);
    }
  };

  useEffect(() => {
    updateOperatorDetails(currentNode.id);
    updateArtifactDetails(currentNode.id);
  }, [currentNode.id, workflow.selectedResult?.id]);

  const onPaneClicked = (event: React.MouseEvent) => {
    event.preventDefault();
    dispatch(setBottomSideSheetOpenState(false));

    // Reset selected node
    dispatch(resetSelectedNode());
  };

  const selectedDag = workflow.selectedDag;
  const getDagDetails = () => {
    if (
      workflow.loadingStatus.loading === LoadingStatusEnum.Succeeded &&
      !!selectedDag
    ) {
      for (const op of Object.values(selectedDag.operators)) {
        // We don't need to call updateArtifactDetails because
        // updateOperatorDetails automatically does that for us.
        updateOperatorDetails(op.id);
      }
    }
  };

  useEffect(getDagDetails, [workflow.selectedDag]);

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
      currentNode.type === NodeType.GenericArtifact
    ) {
      return selectedDag.artifacts[currentNode.id].name;
    } else {
      return selectedDag.operators[currentNode.id].name;
    }
  };

  const getNodeActionButton = () => {
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
        <Button
          onClick={() =>
            exportCsv(artifactResultData, getNodeLabel().replaceAll(' ', '_'))
          }
        >
          Export CSV
        </Button>
      );
    }

    return null;
  };

  return (
    <Layout user={user} layoutType="workspace">
      <Box
        sx={{
          display: 'flex',
          width: '100%',
          height: '100%',
          flexDirection: 'column',
        }}
      >
        {workflow.selectedDag && (
          <WorkflowHeader
            user={user}
            workflowDag={workflow.selectedDag}
            workflowId={workflowId}
          />
        )}

        <Divider />

        <Box
          sx={{
            flex: 1,
            mt: 2,
            p: 3,
            mb: 0,
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
        <Drawer anchor="right" variant="persistent" open={true}>
          <Box width="800px" maxWidth="800px" minHeight="80vh">
            <Box
              width="100%"
              sx={{ backgroundColor: theme.palette.gray['100'] }}
              display="flex"
            >
              <Box
                sx={{ cursor: 'pointer', m: 1, alignSelf: 'center' }}
                onClick={onPaneClicked}
              >
                <FontAwesomeIcon icon={faChevronRight} />
              </Box>
              <Typography variant="h5" padding="16px">
                {getNodeLabel()}
              </Typography>
              <Box sx={{ mx: 2, alignSelf: 'center', marginLeft: 'auto' }}>
                {getNodeActionButton()}
              </Box>
            </Box>
            <Box marginLeft="16px">
              {getDataSideSheetContent(user, currentNode)}
            </Box>
          </Box>
        </Drawer>
      )}

      <WorkflowStatusBar user={user} />
    </Layout>
  );
};

export default WorkflowPage;
