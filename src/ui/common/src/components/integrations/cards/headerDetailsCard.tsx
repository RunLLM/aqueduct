import { faEllipsis } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Tooltip } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import { Integration } from '../../../utils/integrations';
import ExecutionStatus from '../../../utils/shared';
import { StatusIndicator } from '../../workflows/workflowStatus';
import IntegrationLogo from '../logo';

type ResourceHeaderDetailsCardProps = {
  integration: Integration;

  // Eg: "Used by 2 workflows"
  numWorkflowsUsingMsg: string;
};

export const ResourceHeaderDetailsCard: React.FC<
  ResourceHeaderDetailsCardProps
> = ({ integration, numWorkflowsUsingMsg }) => {
  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        width: '900px',
      }}
    >
      <Box display="flex" flexDirection="row" alignItems="center">
        <IntegrationLogo
          service={integration.service}
          size="medium"
          activated
        />

        <Box display="flex" flexDirection="column" sx={{ ml: 2, mr: 2 }}>
          <Box display="flex" flexDirection="row" alignItems={'center'}>
            <Typography sx={{ fontWeight: 400, mr: 2 }} variant="h5">
              {integration.name}
            </Typography>

            <StatusIndicator
              status={
                integration.exec_state?.status || ExecutionStatus.Succeeded
              }
              size="20px"
            />

            <Box sx={{ ml: 2 }}>
              <Tooltip title="See more" arrow>
                <FontAwesomeIcon
                  icon={faEllipsis}
                  style={{
                    transition: 'transform 200ms',
                  }}
                />
              </Tooltip>
            </Box>
          </Box>

          <Typography variant="caption" sx={{ fontWeight: 300}}>
            {new Date(integration.createdAt * 1000).toLocaleString()}
          </Typography>

          <Typography variant="body2" sx={{ fontWeight: 300}}>
            {numWorkflowsUsingMsg}
          </Typography>
        </Box>
      </Box>
    </Box>
  );
};
