import { Typography } from '@mui/material';
import React from 'react';

import { stringToExecutionStatus } from '../../utils/shared';
import { StatusIndicator } from '../workflows/workflowStatus';

interface MetricTableItemProps {
  metricValue?: string;
  // ExecutionStatus serialized as a string.
  status?: string;
}

export const MetricTableItem: React.FC<MetricTableItemProps> = ({
  metricValue,
  status,
}) => {
  if (!metricValue) {
    return (
      <StatusIndicator
        status={stringToExecutionStatus(status)}
        size={'16px'}
        monochrome={false}
      />
    );
  }

  return (
    <Typography
      sx={{
        fontWeight: 300,
      }}
    >
      {metricValue.toString()}
    </Typography>
  );
};

export default MetricTableItem;
