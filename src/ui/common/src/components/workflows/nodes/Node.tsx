import { IconDefinition } from '@fortawesome/fontawesome-svg-core';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';
import { Handle, Position } from 'react-flow-renderer';
import { useSelector } from 'react-redux';

import { RootState } from '../../../stores/store';
import { theme } from '../../../styles/theme/theme';
import { ReactFlowNodeData, ReactflowNodeType } from '../../../utils/reactflow';
import ExecutionStatus, { ExecState, FailureType } from '../../../utils/shared';
import { BaseNode } from './BaseNode.styles';

type Props = {
  data: ReactFlowNodeData;
  defaultLabel: string;
  isConnectable: boolean;
  icon?: IconDefinition;
};

export const Node: React.FC<Props> = ({
  data,
  defaultLabel,
  isConnectable,
  icon,
}) => {
  const label = data.label ? data.label : defaultLabel;
  const currentNode = useSelector(
    (state: RootState) => state.nodeSelectionReducer.selected
  );
  const workflowState = useSelector(
    (state: RootState) => state.workflowReducer
  );
  const selected = currentNode.id === data.nodeId;

  let execState: ExecState;
  if (data.nodeType === ReactflowNodeType.Operator) {
    execState = workflowState.operatorResults[data.nodeId]?.result?.exec_state;
  } else {
    execState = workflowState.artifactResults[data.nodeId]?.result?.exec_state;
  }

  const textColor = selected
    ? theme.palette.DarkContrast50
    : theme.palette.DarkContrast;
  const borderColor = textColor;

  let backgroundColor, hoverColor;
  if (execState?.status === ExecutionStatus.Succeeded) {
    backgroundColor = selected
      ? theme.palette.DarkSuccessMain50
      : theme.palette.DarkSuccessMain;
    hoverColor = theme.palette.DarkSuccessMain75;

    // Warning color for non-fatal errors.
  } else if (
    execState?.status === ExecutionStatus.Failed &&
    execState.failure_type == FailureType.UserNonFatal
  ) {
    backgroundColor = selected
      ? theme.palette.DarkWarningMain50
      : theme.palette.DarkWarningMain;
    hoverColor = theme.palette.DarkWarningMain75;
  } else if (execState?.status === ExecutionStatus.Failed) {
    backgroundColor = selected
      ? theme.palette.DarkErrorMain50
      : theme.palette.DarkErrorMain;
    hoverColor = theme.palette.DarkErrorMain75;
  } else if (execState?.status === ExecutionStatus.Canceled) {
    backgroundColor = selected ? 'gray.700' : 'gray.500';
    hoverColor = 'gray.600';
  } else if (execState?.status === ExecutionStatus.Pending) {
    backgroundColor = selected ? 'blue.300' : 'blue.100';
    hoverColor = 'blue.200';
  }

  return (
    <BaseNode
      sx={{
        backgroundColor,
        color: textColor,
        borderColor: borderColor,
        '&:hover': { backgroundColor: hoverColor },
      }}
    >
      {icon && (
        <Box sx={{ fontSize: '50px', mb: '2px' }}>
          <FontAwesomeIcon icon={icon} />
        </Box>
      )}

      <Typography
        sx={{
          fontSize: '18px',
          maxWidth: '200px',
          minWidth: '140px',
          overflow: 'clip',
          overflowWrap: 'break-word',
          textAlign: 'center',
        }}
      >
        {label}
      </Typography>

      <Handle
        type="source"
        id="db-source-id"
        style={{
          background: theme.palette.darkGray as string,
          border: theme.palette.darkGray as string,
        }}
        isConnectable={isConnectable}
        position={Position.Right}
      />
      <Handle
        type="target"
        id="db-target-id"
        style={{
          background: theme.palette.darkGray as string,
          border: theme.palette.darkGray as string,
        }}
        isConnectable={isConnectable}
        position={Position.Left}
      />
    </BaseNode>
  );
};

export default Node;
