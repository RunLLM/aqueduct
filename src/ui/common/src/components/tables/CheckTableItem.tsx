import React from 'react';

import ExecutionStatus from '../../utils/shared';
import { StatusIndicator } from '../workflows/workflowStatus';

interface CheckTableItemProps {
  status: ExecutionStatus;
  value?: string;
}

export const CheckTableItem: React.FC<CheckTableItemProps> = ({
  status,
  value,
}) => {
  if (value) {
    return <>{value}</>;
  }

  if (status) {
    return <StatusIndicator status={status} size={'16px'} monochrome={false} />;
  }

  return (
    <StatusIndicator
      status={ExecutionStatus.Unknown}
      size={'16px'}
      monochrome={false}
    />
  );
};

export default CheckTableItem;
