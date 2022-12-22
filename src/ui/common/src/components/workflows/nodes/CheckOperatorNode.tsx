import {
  faCheck,
  faCircleCheck,
  faExclamation,
  faMagnifyingGlass,
  faSkullCrossbones,
  faXmark,
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

type Props = {
  data: ReactFlowNodeData;
  isConnectable: boolean;
};

export const checkOperatorNodeIcon = faMagnifyingGlass;

const CheckOperatorNode: React.FC<Props> = ({ data, isConnectable }) => {
  const defaultLabel = 'Check';
  const label = data.label ? data.label : defaultLabel;
  const result = data.result;
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

  let backgroundColor, hoverColor, displayValue;
  if (execState?.status === ExecutionStatus.Succeeded) {
    backgroundColor = selected
      ? theme.palette.DarkSuccessMain50
      : theme.palette.DarkSuccessMain;
    hoverColor = theme.palette.DarkSuccessMain75;

    const icon = result === 'true' ? faCheck : faExclamation;

    displayValue = (
      <>
        <Box sx={{ fontSize: '24px', marginRight: '8px' }}>
          <FontAwesomeIcon icon={icon} />
        </Box>
        <Typography variant="body1">
          {data.result === 'true' ? 'passed' : 'failed'}
        </Typography>
      </>
    );

    // Warning color for non-fatal errors.
  } else if (
    execState?.status === ExecutionStatus.Failed &&
    execState.failure_type === FailureType.UserNonFatal
  ) {
    backgroundColor = selected
      ? theme.palette.DarkWarningMain50
      : theme.palette.DarkWarningMain;
    hoverColor = theme.palette.DarkWarningMain75;

    displayValue = (
      <>
        <Box sx={{ fontSize: '50px' }}>
          <FontAwesomeIcon icon={faSkullCrossbones} />
        </Box>
      </>
    );
  } else if (execState?.status === ExecutionStatus.Failed) {
    backgroundColor = selected
      ? theme.palette.DarkErrorMain50
      : theme.palette.DarkErrorMain;
    hoverColor = theme.palette.DarkErrorMain75;

    displayValue = (
      <>
        <Box sx={{ fontSize: '50px' }}>
          <FontAwesomeIcon icon={faSkullCrossbones} />
        </Box>
      </>
    );
  } else if (execState?.status === ExecutionStatus.Canceled) {
    backgroundColor = selected ? 'gray.700' : 'gray.500';
    hoverColor = 'gray.600';

    displayValue = (
      <>
        <Box sx={{ fontSize: '50px' }}>
          <FontAwesomeIcon icon={faXmark} />
        </Box>
      </>
    );
  } else if (execState?.status === ExecutionStatus.Pending) {
    backgroundColor = selected ? 'blue.300' : 'blue.100';
    hoverColor = 'blue.200';

    displayValue = (
      <>
        <Typography variant="body1" sx={{ fontSize: '25px' }}>
          Pending...
        </Typography>
      </>
    );
  }

  return (
    <BaseNode
      sx={{
        backgroundColor,
        color: textColor,
        borderColor: borderColor,
        '&:hover': { backgroundColor: hoverColor },
        minHeight: 'unset',
        minWidth: '240px',
        padding: '0px',
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
            <FontAwesomeIcon icon={faCircleCheck} />
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
        {displayValue}
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

export default memo(CheckOperatorNode);
