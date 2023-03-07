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
import { StatusIndicator } from '../workflowStatus';
import { BaseNode } from './BaseNode.styles';

type Props = {
  data: ReactFlowNodeData;
  defaultLabel: string;
  isConnectable: boolean;
  icon?: IconDefinition;
  statusLabels: { [key: string]: string };
  // The preview is only shown if the status of this node is succeeded.
  // If it is, then we replace the label with the preview. If the preview
  // is null or the status is not succeeded, then we show the regular label.
  preview?: string;
};

const iconFontSize = '32px';

export const Node: React.FC<Props> = ({
  data,
  defaultLabel,
  isConnectable,
  icon,
  statusLabels,
  preview = null,
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

  if (!execState || !execState.status) {
    return null;
  }

  const textColor = selected
    ? theme.palette.DarkContrast50
    : theme.palette.DarkContrast;
  const borderColor = textColor;

  let status = execState?.status;
  if (
    execState?.status === ExecutionStatus.Failed &&
    execState.failure_type == FailureType.UserNonFatal
  ) {
    status = ExecutionStatus.Warning;
  }

  let backgroundColor;
  if (execState?.status === ExecutionStatus.Succeeded) {
    backgroundColor = theme.palette.green[100];
  } else if (
    execState?.status === ExecutionStatus.Failed &&
    execState.failure_type == FailureType.UserNonFatal
  ) {
    backgroundColor = theme.palette.orange[100];
  } else if (execState?.status === ExecutionStatus.Failed) {
    backgroundColor = theme.palette.red[100];
  } else if (execState?.status === ExecutionStatus.Canceled) {
    backgroundColor = theme.palette.gray[200];
  } else if (execState?.status === ExecutionStatus.Pending) {
    backgroundColor = theme.palette.gray[200];
  }

  const statusIndicatorComponent = (
    <Box
      sx={{
        display: 'flex',
        alignItems: 'center',
        backgroundColor: backgroundColor,
        borderBottomRightRadius: '8px',
        borderBottomLeftRadius: '8px',
      }}
      flex={1}
      height="50%"
      width="100%"
    >
      <Box ml={1}>
        <StatusIndicator
          status={status}
          size={iconFontSize}
          includeTooltip={false}
        />
      </Box>

      <Typography ml={1} textTransform="capitalize" fontSize="28px">
        {preview ?? statusLabels[status]}
      </Typography>
    </Box>
  );
  return (
    <BaseNode
      sx={{
        color: textColor,
        borderColor: borderColor,
      }}
    >
      <Box
        display="flex"
        flexDirection="column"
        alignItems="start"
        width="100%"
        height="100%"
      >
        <Box
          display="flex"
          alignItems="center"
          width="100%"
          height="50%"
          flex={1}
          sx={{
            backgroundColor: theme.palette.gray[300],
            borderTopLeftRadius: '8px',
            borderTopRightRadius: '8px',
          }}
        >
          {icon && (
            <Box sx={{ ml: 1, mr: 2, fontSize: iconFontSize }}>
              <FontAwesomeIcon icon={icon} />
            </Box>
          )}

          <Typography
            sx={{
              maxWidth: '80%',
              flex: 1,
              whiteSpace: 'nowrap',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              fontSize: '32px',
            }}
          >
            {label}
          </Typography>
        </Box>

        {statusIndicatorComponent}
      </Box>

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
