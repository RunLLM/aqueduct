import {
  faChevronDown,
  faChevronUp,
  faCircleCheck,
  faCircleExclamation,
  faCircleInfo,
  faTriangleExclamation,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import {
  NodeType,
  OperatorTypeToNodeTypeMap,
  selectNode,
} from '../../reducers/nodeSelection';
import {
  setBottomSideSheetOpenState,
  setRightSideSheetOpenState,
  setWorkflowStatusBarOpenState,
} from '../../reducers/openSideSheet';
import {
  ArtifactResult,
  handleGetArtifactResults,
  handleGetOperatorResults,
  OperatorResult,
} from '../../reducers/workflow';
import { AppDispatch, RootState } from '../../stores/store';
import { theme } from '../../styles/theme/theme';
import { Artifact } from '../../utils/artifacts';
import UserProfile from '../../utils/auth';
import { Operator } from '../../utils/operators';
import ExecutionStatus, { ExecState, FailureType } from '../../utils/shared';
import getUniqueListBy from '../utils/list_utils';

enum WorkflowStatusTabs {
  Errors = 'ERRORS',
  Logs = 'LOGS',
  Warnings = 'WARNINGS',
  Checks = 'CHECKS',
  Collapsed = 'COLLAPSED',
}

type WorkflowStatusItem = {
  id: string;
  level: WorkflowStatusTabs;
  title: string;
  message: string;
  nodeId: string;
  type: string;
};

interface ActiveWorkflowStatusTabProps {
  setActiveWorkflowStatusTab: (tab: WorkflowStatusTabs) => void;
  activeWorkflowStatusTab: string;
  listItems: WorkflowStatusItem[];
}

export const StatusBarHeaderHeightInPx = 41;
export const CollapsedStatusBarWidthInPx = 75;
export const StatusBarWidthInPx = 432;
export const MaxStatusBarListHeightInPx = 800;

const ActiveWorkflowStatusTab: React.FC<ActiveWorkflowStatusTabProps> = ({
  activeWorkflowStatusTab,
  listItems,
  setActiveWorkflowStatusTab,
}) => {
  const openSideSheetState = useSelector(
    (state: RootState) => state.openSideSheetReducer
  );
  const dispatch: AppDispatch = useDispatch();

  const workflowStatusIcons = {
    [WorkflowStatusTabs.Errors]: (
      <Box sx={{ fontSize: '20px', color: theme.palette.red['500'] }}>
        <FontAwesomeIcon icon={faCircleExclamation} />
      </Box>
    ),
    [WorkflowStatusTabs.Logs]: (
      <Box sx={{ fontSize: '20px', color: theme.palette.blue['400'] }}>
        <FontAwesomeIcon icon={faCircleInfo} />
      </Box>
    ),
    [WorkflowStatusTabs.Warnings]: (
      <Box sx={{ fontSize: '20px', color: theme.palette.orange['500'] }}>
        <FontAwesomeIcon icon={faTriangleExclamation} />
      </Box>
    ),
    [WorkflowStatusTabs.Checks]: (
      <Box sx={{ fontSize: '20px', color: theme.palette.green['400'] }}>
        <FontAwesomeIcon icon={faCircleCheck} />
      </Box>
    ),
  };

  if (
    activeWorkflowStatusTab === WorkflowStatusTabs.Collapsed ||
    !openSideSheetState.workflowStatusBarOpen
  ) {
    return null;
  }

  /**
   * This function takes in a dispatch call (which must be created in a
   * component) and a call to a set state function in the using component, and it
   * returns a function which takes an event for a click on a node in ReactFlow
   * and opens the appropriate corresponding sidesheet.
   */
  const switchSideSheet = (nodeId: string, type: string) => {
    dispatch(selectNode({ id: nodeId, type: type as NodeType }));
    dispatch(setRightSideSheetOpenState(true));
    dispatch(setBottomSideSheetOpenState(true));
  };

  return (
    <Box
      sx={{
        width: `${StatusBarWidthInPx}px`,
        maxHeight: `${MaxStatusBarListHeightInPx}px`,
        position: 'absolute',
        overflow: 'auto',
        backgroundColor: 'white',
        borderRadius: '8px',
        zIndex: 10,
        border: `1px solid`,
        borderColor: 'gray.500',
      }}
    >
      {listItems.map((listItem, index) => {
        const key = listItem.nodeId.length > 0 ? listItem.nodeId : index;
        return (
          <Box
            key={key}
            sx={{
              display: 'flex',
              flexDirection: 'row',
              width: '100%',
              backgroundColor: 'white',
              borderBottom: index === listItems.length - 1 ? null : `1px solid`,
              borderColor: 'gray.500',
              alignItems: 'start',
            }}
          >
            <Box sx={{ marginLeft: '8px', marginTop: '16px' }}>
              {listItem ? workflowStatusIcons[listItem.level] : null}
            </Box>
            <Box
              sx={{
                display: 'flex',
                flexDirection: 'column',
                verticalAlign: 'middle',
                padding: 2,
              }}
            >
              <Typography
                sx={{
                  fontFamily: 'Monospace',
                  fontWeight: 'bold',
                  marginRight: 2,
                  whiteSpace: 'normal',
                  '&:hover': { textDecoration: 'underline', cursor: 'pointer' },
                }}
                onClick={() => {
                  if (listItem.nodeId.length > 0 && listItem.type.length > 0) {
                    switchSideSheet(listItem.nodeId, listItem.type);
                    dispatch(setWorkflowStatusBarOpenState(false));
                    setActiveWorkflowStatusTab(WorkflowStatusTabs.Collapsed);
                  }
                }}
              >
                {listItem.title}
              </Typography>
              <Typography
                sx={{
                  fontFamily: 'Monospace',
                  fontWeight: 'light',
                  marginTop: '2px',
                  fontSize: '12px',
                  whiteSpace: 'pre-wrap',
                }}
              >
                {listItem.message}
              </Typography>
            </Box>
          </Box>
        );
      })}
    </Box>
  );
};

type WorkflowStatusBarProps = {
  user: UserProfile;
};

export const WorkflowStatusBar: React.FC<WorkflowStatusBarProps> = ({
  user,
}) => {
  const dispatch: AppDispatch = useDispatch();
  const workflow = useSelector((state: RootState) => state.workflowReducer);
  const selectedDag = workflow.selectedDag;
  const artifacts: { [id: string]: Artifact } =
    workflow.selectedDag?.artifacts ?? {};
  const operators: { [id: string]: Operator } =
    workflow.selectedDag?.operators ?? {};

  const [activeWorkflowStatusTab, setActiveWorkflowStatusTab] = useState(
    WorkflowStatusTabs.Collapsed
  );

  const [numErrors, setNumErrors] = useState(0);
  const [numWarnings, setNumWarnings] = useState(0);
  const [numWorkflowLogs, setNumWorkflowLogs] = useState(0);
  const [numWorkflowChecksPassed, setNumWorkflowChecksPassed] = useState(0);

  // List of all the workflow status items
  const [workflowStatusItems, setWorkflowStatusItems] = useState<
    WorkflowStatusItem[]
  >([]);

  // List of the workflow status items filtered out by category: errors, warnings, logs and checks passed.
  const [listItems, setListItems] = useState<WorkflowStatusItem[]>([]);

  useEffect(() => {
    setWorkflowStatusItems(normalizeWorkflowStatusItems());
  }, [workflow, selectedDag, artifacts, operators]); // recompute state when all derived values change.

  useEffect(() => {
    const filteredErrors: WorkflowStatusItem[] = workflowStatusItems.filter(
      (workflowStatusItem) => {
        return workflowStatusItem.level === WorkflowStatusTabs.Errors;
      }
    );

    const filteredWarnings: WorkflowStatusItem[] = workflowStatusItems.filter(
      (workflowStatusItem) => {
        return workflowStatusItem.level === WorkflowStatusTabs.Warnings;
      }
    );

    const filteredLogs: WorkflowStatusItem[] = workflowStatusItems.filter(
      (workflowStatusItem) => {
        return workflowStatusItem.level === WorkflowStatusTabs.Logs;
      }
    );

    const filteredChecks: WorkflowStatusItem[] = workflowStatusItems.filter(
      (workflowStatusItem) => {
        return workflowStatusItem.level === WorkflowStatusTabs.Checks;
      }
    );

    switch (activeWorkflowStatusTab) {
      case WorkflowStatusTabs.Errors: {
        if (filteredErrors.length === 0) {
          setListItems([
            {
              id: '1',
              level: WorkflowStatusTabs.Errors,
              title: 'No errors.',
              message: '',
              nodeId: '',
              type: '',
            },
          ]);
        } else {
          setListItems(filteredErrors);
        }
        break;
      }
      case WorkflowStatusTabs.Warnings: {
        if (filteredWarnings.length === 0) {
          setListItems([
            {
              id: '1',
              level: WorkflowStatusTabs.Warnings,
              title: 'No warnings.',
              message: '',
              nodeId: '',
              type: '',
            },
          ]);
        } else {
          setListItems(filteredWarnings);
        }
        break;
      }
      case WorkflowStatusTabs.Logs: {
        if (filteredLogs.length === 0) {
          setListItems([
            {
              id: '1',
              level: WorkflowStatusTabs.Logs,
              title: 'No logs.',
              message: '',
              nodeId: '',
              type: '',
            },
          ]);
        } else {
          setListItems(filteredLogs);
        }
        break;
      }
      case WorkflowStatusTabs.Checks: {
        if (filteredChecks.length === 0) {
          setListItems([
            {
              id: '1',
              level: WorkflowStatusTabs.Checks,
              title: 'No successful results.',
              message: '',
              nodeId: '',
              type: '',
            },
          ]);
        } else {
          setListItems(filteredChecks);
        }
        break;
      }
      default: {
        setListItems(filteredErrors);
        break;
      }
    }

    setNumErrors(filteredErrors.length);
    setNumWarnings(filteredWarnings.length);
    setNumWorkflowLogs(filteredLogs.length);
    setNumWorkflowChecksPassed(filteredChecks.length);
  }, [workflowStatusItems, activeWorkflowStatusTab]);

  const selectTab = (tab: WorkflowStatusTabs) => {
    dispatch(setWorkflowStatusBarOpenState(true));
    setActiveWorkflowStatusTab(tab);
  };

  const collapseWorkflowStatusBar = (event: React.MouseEvent) => {
    event.stopPropagation();
    dispatch(setWorkflowStatusBarOpenState(false));
    setActiveWorkflowStatusTab(WorkflowStatusTabs.Collapsed);
  };

  // Chevron up is clicked. Errors tab is left most tab, so we select that one.
  const expandWorkflowStatusbar = (event: React.MouseEvent) => {
    event.stopPropagation();
    selectTab(WorkflowStatusTabs.Errors);
  };

  const normalizeWorkflowStatusItems = () => {
    const normalizedWorkflowStatusItems: WorkflowStatusItem[] = [];

    Object.keys(artifacts).map(async (artifactId) => {
      const artifactName = artifacts[artifactId].name
        ? artifacts[artifactId].name
        : 'Artifact';
      const artifactResult: ArtifactResult =
        workflow.artifactResults[artifactId];

      // Check if artifactResult is in the map, if not fetch it.
      if (!artifactResult) {
        dispatch(
          handleGetArtifactResults({
            apiKey: user.apiKey,
            workflowDagResultId: workflow.selectedResult.id,
            artifactId: artifactId,
          })
        );

        return;
      }

      const newWorkflowStatusItem: WorkflowStatusItem = {
        id: `id-artifactResult-${artifactId}`,
        level: WorkflowStatusTabs.Checks,
        title: `${artifactName} Failed`,
        message: '',
        nodeId: artifactId,
        type: 'tableArtifact',
      };

      const artifactStatus: ExecutionStatus = artifactResult.result?.status;
      const artifactExecState: ExecState = artifactResult.result?.exec_state;

      if (
        artifactStatus === ExecutionStatus.Failed &&
        artifactExecState.failure_type == FailureType.UserNonFatal
      ) {
        newWorkflowStatusItem.level = WorkflowStatusTabs.Warnings;
        newWorkflowStatusItem.title = `Non-fatal error occurred for ${artifactName}`;
        newWorkflowStatusItem.message = artifactExecState.error?.tip;
      } else if (artifactStatus === ExecutionStatus.Failed) {
        newWorkflowStatusItem.level = WorkflowStatusTabs.Errors;
        newWorkflowStatusItem.title = `Error creating ${artifactName}.`;
        newWorkflowStatusItem.message = `Unable to create artifact ${artifactName} (${artifactId}).`;
      } else if (artifactStatus === ExecutionStatus.Succeeded) {
        newWorkflowStatusItem.level = WorkflowStatusTabs.Checks;
        newWorkflowStatusItem.title = `Artifact ${artifactName} created.`;
        newWorkflowStatusItem.message = `Successfully created artifact ${artifactName} (${artifactId})`;
      } else {
        // artifact is still pending, skip adding to list of workflow status items.
        return;
      }

      // add workflow status item to the list.
      normalizedWorkflowStatusItems.push(newWorkflowStatusItem);
    });

    Object.keys(operators).map(async (operatorId) => {
      const operatorName = operators[operatorId].name
        ? operators[operatorId].name
        : 'Operator';

      const operatorResult: OperatorResult =
        workflow.operatorResults[operatorId];

      if (!operatorResult) {
        // We can fetch it here, or we can also do so when user opens the respective node.
        // Not sure which time is the best to do this.
        dispatch(
          handleGetOperatorResults({
            apiKey: user.apiKey,
            workflowDagResultId: workflow.selectedResult.id,
            operatorId: operatorId,
          })
        );
        return;
      }

      const newWorkflowStatusItem: WorkflowStatusItem = {
        id: `id-operatorResult-${operatorId}`,
        level: WorkflowStatusTabs.Checks,
        title: `${operatorName} Failed`,
        message: '',
        nodeId: operatorId,
        type: OperatorTypeToNodeTypeMap[
          operators[operatorId].spec.type
        ].toString(),
      };

      const opExecState: ExecState = operatorResult.result?.exec_state;
      const operatorExecutionStatus: ExecutionStatus = operatorResult.result
        ? operatorResult.result.exec_state.status
        : null;

      if (
        operatorExecutionStatus === ExecutionStatus.Failed &&
        opExecState.failure_type === FailureType.UserNonFatal
      ) {
        newWorkflowStatusItem.level = WorkflowStatusTabs.Warnings;
        newWorkflowStatusItem.title = `Warning for ${operatorName}`;
        newWorkflowStatusItem.message = opExecState.error?.tip;
      } else if (operatorExecutionStatus === ExecutionStatus.Failed) {
        // add to the errors array.
        newWorkflowStatusItem.level = WorkflowStatusTabs.Errors;
        if (!!opExecState.error) {
          newWorkflowStatusItem.title = `Error executing ${operatorName} (${operatorId})`;
          const err = opExecState.error;
          newWorkflowStatusItem.message = `${err.tip ?? ''}\n${
            err.context ?? ''
          }`;
        } else {
          // no error message found, so treat this as a system internal error
          newWorkflowStatusItem.message = `Aqueduct Internal Error`;
        }
      } else if (operatorExecutionStatus === ExecutionStatus.Succeeded) {
        newWorkflowStatusItem.level = WorkflowStatusTabs.Checks;
        newWorkflowStatusItem.title = `${operatorName} succeeded`;
        newWorkflowStatusItem.message = `Operator successfully executed`;
      } else {
        // operator result is still pending, so skip current item since we do not know if successful or failed.
        return;
      }

      // add workflow status item to the list.
      normalizedWorkflowStatusItems.push(newWorkflowStatusItem);

      // LEFT off here, see the normalize logs function and work from there :)
      if (opExecState && opExecState.user_logs) {
        const logs = opExecState?.user_logs;
        const stdoutLines = (logs.stdout ?? '').split('\n');
        for (let i = 0; i < stdoutLines.length - 1; i++) {
          normalizedWorkflowStatusItems.push({
            id: `${operatorId}-stdout-${i}`,
            level: WorkflowStatusTabs.Logs,
            title: `${operatorName} stdout`,
            message: stdoutLines[i],
            nodeId: operatorId,
            type: OperatorTypeToNodeTypeMap[
              operators[operatorId].spec.type
            ].toString(),
          });
        }

        const stderrLines = (logs.stderr ?? '').split('\n');
        for (let i = 0; i < stderrLines.length - 1; i++) {
          normalizedWorkflowStatusItems.push({
            id: `${operatorId}-stderr-${i}`,
            level: WorkflowStatusTabs.Logs,
            title: `${operatorName} stderr`,
            message: stderrLines[i],
            nodeId: operatorId,
            type: OperatorTypeToNodeTypeMap[
              operators[operatorId].spec.type
            ].toString(),
          });
        }
      }
    });

    return getUniqueListBy<WorkflowStatusItem>(
      normalizedWorkflowStatusItems,
      'id'
    );
  };

  const collapsed = activeWorkflowStatusTab === WorkflowStatusTabs.Collapsed;

  const statusBarIconStyles = {
    mx: 1,
    py: 1,
    width: '40px',
    cursor: 'pointer',
    alignItems: 'start',
    display: 'flex',
  };

  return (
    <Box
      sx={{
        cursor: 'pointer',
        zIndex: 10,
        border: `1px solid ${theme.palette.gray['500']}`,
        borderRadius: '8px',
      }}
      onClick={() => {
        selectTab(WorkflowStatusTabs.Errors);
      }}
    >
      <Box
        sx={{
          display: 'flex',
          flexDirection: 'row',
          alignItems: 'center',
          px: 0,
          ml: 0,
          height: `${StatusBarHeaderHeightInPx}px`,
          borderBottom: null,
          overflowY: 'none',
        }}
      >
        <Box
          onClick={(event: React.MouseEvent) => {
            // handle event here and keep from being handled by root onClick listener of parent div.
            event.stopPropagation();
            selectTab(WorkflowStatusTabs.Errors);
          }}
          sx={{
            ...statusBarIconStyles,
            color:
              activeWorkflowStatusTab === WorkflowStatusTabs.Errors
                ? theme.palette.red['600']
                : theme.palette.red['500'],
            borderBottom:
              activeWorkflowStatusTab === WorkflowStatusTabs.Errors
                ? `2px solid ${theme.palette.red['600']}`
                : '', // red600
            '&:hover': { color: theme.palette.red['600'] },
            fontSize: '20px',
            marginRight: 2,
            marginLeft: 2,
          }}
        >
          <FontAwesomeIcon icon={faCircleExclamation} />
          <Typography sx={{ ml: 1 }}>{numErrors}</Typography>
        </Box>

        <Box
          onClick={(event: React.MouseEvent) => {
            event.stopPropagation();
            selectTab(WorkflowStatusTabs.Warnings);
          }}
          sx={{
            ...statusBarIconStyles,
            color:
              activeWorkflowStatusTab === WorkflowStatusTabs.Warnings
                ? theme.palette.orange['600']
                : theme.palette.orange['500'],
            borderBottom:
              activeWorkflowStatusTab === WorkflowStatusTabs.Warnings
                ? `2px solid ${theme.palette.orange['600']}`
                : '', // orange600
            '&:hover': { color: theme.palette.orange['600'] },
            fontSize: '20px',
            marginRight: 2,
          }}
        >
          <FontAwesomeIcon icon={faTriangleExclamation} />
          <Typography sx={{ ml: 1 }}>{numWarnings}</Typography>
        </Box>

        <Box
          onClick={(event: React.MouseEvent) => {
            event.stopPropagation();
            selectTab(WorkflowStatusTabs.Logs);
          }}
          sx={{
            ...statusBarIconStyles,
            color:
              activeWorkflowStatusTab === WorkflowStatusTabs.Logs
                ? theme.palette.blue['500']
                : theme.palette.blue['400'],
            borderBottom:
              activeWorkflowStatusTab === WorkflowStatusTabs.Logs
                ? `2px solid ${theme.palette.blue['500']}`
                : '', // blue500
            '&:hover': { color: theme.palette.blue['500'] },
            fontSize: '20px',
            marginRight: 2,
          }}
        >
          <FontAwesomeIcon icon={faCircleInfo} />
          <Typography sx={{ ml: 1 }}>{numWorkflowLogs}</Typography>
        </Box>

        <Box
          onClick={(event: React.MouseEvent) => {
            event.stopPropagation();
            selectTab(WorkflowStatusTabs.Checks);
          }}
          sx={{
            ...statusBarIconStyles,
            color:
              activeWorkflowStatusTab === WorkflowStatusTabs.Checks
                ? theme.palette.green['500']
                : theme.palette.green['400'],
            borderBottom:
              activeWorkflowStatusTab === WorkflowStatusTabs.Checks
                ? `2px solid ${theme.palette.green['500']}`
                : '', // green500
            '&:hover': { color: theme.palette.green['500'] },
            fontSize: '20px',
            marginRight: 2,
          }}
        >
          <FontAwesomeIcon icon={faCircleCheck} />
          <Typography sx={{ ml: 1 }}>{numWorkflowChecksPassed}</Typography>
        </Box>

        <Box
          sx={{ cursor: 'pointer', my: 2, marginLeft: 'auto', marginRight: 2 }}
        >
          {collapsed ? (
            <FontAwesomeIcon
              icon={faChevronDown}
              onClick={expandWorkflowStatusbar}
            />
          ) : (
            <FontAwesomeIcon
              icon={faChevronUp}
              onClick={collapseWorkflowStatusBar}
            />
          )}
        </Box>
      </Box>

      <ActiveWorkflowStatusTab
        activeWorkflowStatusTab={activeWorkflowStatusTab}
        listItems={listItems}
        setActiveWorkflowStatusTab={setActiveWorkflowStatusTab}
      />
    </Box>
  );
};

export default WorkflowStatusBar;
