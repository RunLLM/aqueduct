import {
  faHashtag,
  faTemperatureHalf,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { memo } from 'react';
import { Handle, Position } from 'react-flow-renderer';
import { useSelector } from 'react-redux';

import { RootState } from '../../../stores/store';
import { theme } from '../../../styles/theme/theme';
import { ReactFlowNodeData } from '../../../utils/reactflow';
import ExecutionStatus, { ExecState, FailureType } from '../../../utils/shared';
import { BaseNode } from './BaseNode.styles';
import { NodeStatusIconography } from './NodeStatusIconography';

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const metricOperatorNodeIcon = faTemperatureHalf;

export const parseMetricResult = (
  metricValue: string,
  sigfigs: number
): string => {
  // Check if the number passed in is a whole number, return that if so.
  const parsedFloat = parseFloat(metricValue);
  if (parsedFloat % 1 === 0) {
    return metricValue;
  }
  // Only show three decimal points.
  return parsedFloat.toFixed(sigfigs);
};

const MetricOperatorNode: React.FC<Props> = ({ data, isConnectable }) => {
  const defaultLabel = 'Metric';
  const label = data.label ? data.label : defaultLabel;
  const currentNode = useSelector(
    (state: RootState) => state.nodeSelectionReducer.selected
  );
  const workflowState = useSelector(
    (state: RootState) => state.workflowReducer
  );
  const selected = currentNode.id === data.nodeId;
  const execState: ExecState =
    workflowState.operatorResults[data.nodeId]?.result?.exec_state;

  const textColor = selected
    ? theme.palette.DarkContrast50
    : theme.palette.DarkContrast;
  const borderColor = textColor;

  const successDisplay = (
    <Typography variant="h5">{parseMetricResult(data.result, 3)}</Typography>
  );

  let backgroundColor, hoverColor;
  if (execState?.status === ExecutionStatus.Succeeded) {
    backgroundColor = selected
      ? theme.palette.DarkSuccessMain50
      : theme.palette.DarkSuccessMain;
    hoverColor = theme.palette.DarkSuccessMain75;
    // Warning color for non-fatal errors.
  } else if (
    execState?.status === ExecutionStatus.Failed &&
    execState.failure_type === FailureType.UserNonFatal
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
        //   minHeight: 'unset',
        //   minWidth: '240px',
        //   padding: '0px',
      }}
    >
      <Box
        sx={{
          height: '32px',
          width: '100%',
          borderBottom: `1px solid ${borderColor}`,
        }}
      >
        <Box
          width="100%"
          height="100%"
          display="flex"
          justifyContent="center"
          alignItems="center"
        >
          <Box sx={{ fontSize: '24px', marginRight: '8px' }}>
            <FontAwesomeIcon icon={faHashtag} />
          </Box>
          <Typography variant="body1">{label}</Typography>
        </Box>
      </Box>

      <Box
        width="100%"
        height="100%"
        minHeight="80px"
        display="flex"
        justifyContent="center"
        alignItems="center"
      >
        <NodeStatusIconography
          execState={execState}
          successDisplay={successDisplay}
        />
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

export default memo(MetricOperatorNode);
