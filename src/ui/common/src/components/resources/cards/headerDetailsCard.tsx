import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import {
  resolveLogoService,
  Resource,
  resourceExecState,
} from '../../../utils/resources';
import { StatusIndicator } from '../../workflows/workflowStatus';
import ResourceLogo from '../logo';

type ResourceHeaderDetailsCardProps = {
  resource: Resource;

  // Eg: "Used by 2 workflows"
  numWorkflowsUsingMsg: string;
};

export const ResourceHeaderDetailsCard: React.FC<
  ResourceHeaderDetailsCardProps
> = ({ resource, numWorkflowsUsingMsg }) => {
  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <Box display="flex" flexDirection="row" alignItems="center">
        <ResourceLogo
          service={resolveLogoService(resource)}
          size="medium"
          activated
        />

        <Box display="flex" flexDirection="column" sx={{ ml: 2, mr: 2 }}>
          <Box display="flex" flexDirection="row" alignItems={'center'}>
            <Typography sx={{ fontWeight: 400, mr: 2 }} variant="h5">
              {resource.name}
            </Typography>

            <StatusIndicator
              status={resourceExecState(resource).status}
              size="20px"
            />
          </Box>

          <Typography variant="caption" sx={{ fontWeight: 300 }}>
            {new Date(resource.createdAt * 1000).toLocaleString()}
          </Typography>

          <Typography variant="body2" sx={{ fontWeight: 300 }}>
            {numWorkflowsUsingMsg}
          </Typography>
        </Box>
      </Box>
    </Box>
  );
};
