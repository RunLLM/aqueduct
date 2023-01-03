import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faSkullCrossbones,
  faXmark,
} from '@fortawesome/free-solid-svg-icons';
import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React from 'react';

import ExecutionStatus, { ExecState, FailureType } from '../../../utils/shared';

type Props = {
  execState: ExecState;
  successDisplay: JSX.Element;
};

export const NodeStatusIconography: React.FC<Props> = ({ execState, successDisplay }) => {
  if (!execState) {
    return (
      <>
        <Typography variant="body1" sx={{ fontSize: '25px' }}>
          Loading
        </Typography>
      </>
    );
  }
  if (execState.status === ExecutionStatus.Succeeded) {
    return successDisplay;
  } else if (
    execState.status === ExecutionStatus.Failed
  ) {
    return (
      <>
        <Box sx={{ fontSize: '50px' }}>
          <FontAwesomeIcon icon={faSkullCrossbones} />
        </Box>
      </>
    );
  } else if (execState.status === ExecutionStatus.Canceled) {
    return (
      <>
        <Box sx={{ fontSize: '50px' }}>
          <FontAwesomeIcon icon={faXmark} />
        </Box>
      </>
    );
  } else if (execState.status === ExecutionStatus.Pending) {
    return (
      <>
        <Typography variant="body1" sx={{ fontSize: '25px' }}>
          Pending...
        </Typography>
      </>
    );
  }

  return (
    <>
      <Typography variant="body1" sx={{ fontSize: '25px' }}>
        Error
      </Typography>
    </>
  );
};

export default NodeStatusIconography;
