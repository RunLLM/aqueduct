import { Box, Typography } from '@mui/material';
import React from 'react';

import ExecutionStatus, { ExecState } from '../../../utils/shared';
import { StatusIndicator } from '../workflowStatus';

type Props = {
  execState: ExecState;
  successDisplay: JSX.Element;
};

  /**
   * Function used to determine what to display for DAG check or metric nodes.
   * The icons are a bit larger than the default StatusIndicator size, monochrome (black), 
   * and are followed by a label. It also handles the state when we haven't fetched execState 
   * yet and displays the check/metric value if the node was successfully executed.
   * @param execState - The operator's execution state, if it exists.
   * @param successDisplay - What to display if the operator was successfully executed.
   * @returns - JSX element to be displayed on the node.
   */
export const NodeStatusIconography: React.FC<Props> = ({
  execState,
  successDisplay,
}) => {
  const iconSize = '24px'
  let status = ExecutionStatus.Pending;
  let statusLabel = "fetching";
  if (execState) {
    status = execState.status;
    statusLabel = execState.status;
  }
  if (status === ExecutionStatus.Succeeded) {
    return successDisplay;
  } else {
    return (
      <Box display="flex" alignItems="center">
        <StatusIndicator
          status={status}
          size={iconSize}
          monochrome={'black'}
        />
        <Typography variant="body1" sx={{pl: 1 }}>
          {statusLabel}
        </Typography>
      </Box>     
    );
  }
};

export default NodeStatusIconography;
