import { faCircleExclamation } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Tooltip, Typography } from '@mui/material';
import React from 'react';
import { useState } from 'react';

import { theme } from '../../../../styles/theme/theme';
import ExecutionStatus from '../../../../utils/shared';
import { parseMetricResult } from '../../../workflows/nodes/MetricOperatorNode';

interface ShowMoreProps {
  totalItems: number;
  numPreviewItems: number;
  expanded: boolean;
  onClick: () => void;
}

const showMoreStyles = {
  fontWeight: 500,
  color: theme.palette.gray['600'],
  cursor: 'pointer',
  '&:hover': { textDecoration: 'underline' },
};

export const ShowMore: React.FC<ShowMoreProps> = ({
  totalItems,
  numPreviewItems,
  expanded,
  onClick,
}) => {
  // handle edge case where there is only one metric to show.
  if (totalItems === 1) {
    return null;
  }

  let prompt = `Show ${totalItems - numPreviewItems} More`;
  if (expanded) {
    prompt = `Show Less`;
  }

  return (
    <Box onClick={onClick}>
      <Typography variant="body2" sx={showMoreStyles}>
        {prompt}
      </Typography>
    </Box>
  );
};

export interface MetricPreview {
  metricId: string;
  name: string;
  value?: string;
  status: ExecutionStatus;
}

interface MetricItemProps {
  metrics: MetricPreview[];
}

const MetricItem: React.FC<MetricItemProps> = ({ metrics }) => {
  const [expanded, setExpanded] = useState<boolean>(false);
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
          alignItems="center"
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
            <Typography variant="body1">
              {parseMetricResult(metrics[i].value, 3)}
            </Typography>
          )}
        </Box>
      );
    }
  }

  const toggleExpanded = () => {
    setExpanded(!expanded);
  };

  // 0.875rem is the size of the ShowMore.
  // Add this additional padding if we don't have more than 1 metric to keep the rows equal size.
  return (
    <Box>
      {metrics.length > 0 ? (
        <Box sx={metrics.length === 1 && {padding:"0.875rem 0 0.875rem 0"}}>
          {metricList}
          <ShowMore
            totalItems={metrics.length}
            numPreviewItems={metricsToShow}
            expanded={expanded}
            onClick={toggleExpanded}
          />
        </Box>
      ) : (
        <Typography sx={{padding:"0.875rem 0 0.875rem 0"}} variant="body1">No metrics.</Typography>
      )}
    </Box>
  );
};

export default MetricItem;
