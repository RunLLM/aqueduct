import { faCircleExclamation } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Tooltip, Typography } from '@mui/material';
import React from 'react';
import { useState } from 'react';

import { theme } from '../../../../styles/theme/theme';
import ExecutionStatus from '../../../../utils/shared';

export interface MetricPreview {
  // used to fetch additional metrics and information to be shown in table.
  // TODO: Consider showing other metric related meta data here.
  metricId: string;
  name: string;
  value?: string;
  status: ExecutionStatus;
}

interface MetricItemProps {
  metrics: MetricPreview[];
}

const MetricItem: React.FC<MetricItemProps> = ({ metrics }) => {
  const [expanded, setExpanded] = useState(false);
  const metricList = [];

  let metricsToShow = metrics.length;
  if (metrics.length > 0) {
    if (!expanded) {
      metricsToShow = 1;
    }
    for (let i = 0; i < metricsToShow; i++) {
      metricList.push(
        <Box
          display="flex"
          key={metrics[i].metricId}
          justifyContent="space-between"
          height="30px"
        >
          <Typography variant="body1" sx={{ fontWeight: 400 }}>
            {metrics[i].name}
          </Typography>
          {metrics[i].status === ExecutionStatus.Failed ? (
            <Tooltip title="Error" placement="bottom" arrow>
              <Box sx={{ fontSize: '20px', color: theme.palette.red['500'] }}>
                <FontAwesomeIcon icon={faCircleExclamation} />
              </Box>
            </Tooltip>
          ) : (
            <Typography variant="body1">{metrics[i].value}</Typography>
          )}
        </Box>
      );
    }
  }

  const toggleExpanded = () => {
    setExpanded(!expanded);
  };

  const showMoreStyles = {
    fontWeight: 500,
    color: theme.palette.gray['600'],
    cursor: 'pointer',
    '&:hover': { textDecoration: 'underline' },
  };

  // TODO: make into a component to share with checks/metrics list
  const showLess = (
    <Box>
      <Typography variant="body2" sx={showMoreStyles} onClick={toggleExpanded}>
        Show Less ...
      </Typography>
    </Box>
  );

  // TODO: make into a component to share with checks/metrics list
  const showMore = (
    <Box>
      <Typography variant="body2" sx={showMoreStyles} onClick={toggleExpanded}>
        Show More ({metrics.length - metricsToShow}) ...
      </Typography>
    </Box>
  );

  return (
    <Box>
      {metrics.length > 0 ? (
        <>
          {metricList}
          {expanded ? showLess : showMore}
        </>
      ) : (
        <Typography variant="body1">No metrics.</Typography>
      )}
    </Box>
  );
};

export default MetricItem;
