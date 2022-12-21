import Chip from '@mui/material/Chip';
import React from 'react';

import ExecutionStatus from '../../utils/shared';

type Props = {
  status: ExecutionStatus;
};

const ExecutionChip: React.FC<Props> = ({ status }) => {
  if (status === ExecutionStatus.Succeeded) {
    return <Chip label="Succeeded" color="success" size="small" />;
  }

  if (status === ExecutionStatus.Failed) {
    return <Chip label="Failed" color="error" size="small" />;
  }

  if (
    status === ExecutionStatus.Pending ||
    status === ExecutionStatus.Running
  ) {
    return <Chip label="In Progress" color="info" size="small" />;
  }

  if (status === ExecutionStatus.Canceled) {
    return <Chip label="Canceled" color="default" size="small" />;
  }

  if (status === ExecutionStatus.Registered) {
    return <Chip label="Pending" color="info" size="small" />;
  }

  return null;
};

export default ExecutionChip;
