import { Box, Typography } from '@mui/material';
import React from 'react';
import { useState } from 'react';

import { theme } from '../../../../styles/theme/theme';

export interface MetricPreview {
  // used to fetch additional metrics and information to be shown in table.
  // TODO: Consider showing other metric related meta data here.
  metricId: string;
  name: string;
  value: string;
}

interface MetricItemProps {
  metrics: MetricPreview[];
}

const MetricItem: React.FC<MetricItemProps> = ({ metrics }) => {
  const [expanded, setExpanded] = useState(false);
  const metricList = [];
  let metricsToShow = metrics.length;
  if (!expanded && metrics.length > 1) {
    metricsToShow = 1;
  }

  for (let i = 0; i < metricsToShow; i++) {
    metricList.push(
      <Box
        display="flex"
        key={metrics[i].metricId}
        justifyContent="space-between"
      >
        <Typography variant="body1" sx={{ fontWeight: 'bold' }}>
          {metrics[i].name}
        </Typography>
        <Typography variant="body1">{metrics[i].value}</Typography>
      </Box>
    );
  }

  const toggleExpanded = () => {
    setExpanded(!expanded);
  };

  const showMoreStyles = {
    fontWeight: 'bold',
    color: theme.palette.gray['700'],
    cursor: 'pointer',
    '&:hover': { textDecoration: 'underline' },
  };

  // TODO: make into a component to share with checks/metrics list
  const showLess = (
    <Box>
      <Typography variant="body1" sx={showMoreStyles} onClick={toggleExpanded}>
        Show Less ...
      </Typography>
    </Box>
  );

  // TODO: make into a component to share with checks/metrics list
  const showMore = (
    <Box>
      <Typography variant="body1" sx={showMoreStyles} onClick={toggleExpanded}>
        Show More ({metrics.length - metricsToShow}) ...
      </Typography>
    </Box>
  );

  return (
    <Box>
      {metricList}
      {expanded ? showLess : showMore}
    </Box>
  );
};

export default MetricItem;
