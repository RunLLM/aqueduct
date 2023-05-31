import Box from '@mui/material/Box';
import React from 'react';

import {
  resolveLogoService,
  Resource,
  resourceExecState,
} from '../../../utils/resources';
import { StatusIndicator } from '../../workflows/workflowStatus';
import ResourceLogo from '../logo';
import { ResourceFieldsDetailsCard } from './resourceFieldsDetailsCard';
import { TruncatedText } from './text';

type ResourceProps = {
  resource: Resource;

  // Eg: "Used by 2 workflows"
  numWorkflowsUsingMsg: string;
};

export const ResourceCard: React.FC<ResourceProps> = ({
  resource,
  numWorkflowsUsingMsg,
}) => {
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Box sx={{ display: 'flex', flexDirection: 'row', alignItems: 'center' }}>
        <StatusIndicator
          status={resourceExecState(resource).status}
          size="16px"
        />

        {/* Subtract the width of the status indicator, padding, and logo respectively. */}
        <Box
          sx={{ mx: 1, flex: 1, maxWidth: `calc(100% - 16px - 16px - 24px)` }}
        >
          <TruncatedText sx={{ fontWeight: 400 }} variant="h6">
            {resource.name}
          </TruncatedText>
        </Box>
        <ResourceLogo
          service={resolveLogoService(resource)}
          size="small"
          activated
        />
      </Box>

      {/*Leave this empty if resource.createdAt isn't set.*/}
      <TruncatedText
        variant="caption"
        marginBottom={1}
        sx={{ fontWeight: 300 }}
      >
        {resource.createdAt
          ? new Date(resource.createdAt * 1000).toLocaleString()
          : '  '}
      </TruncatedText>

      <ResourceFieldsDetailsCard resource={resource} detailedView={false} />

      <Box
        sx={{
          position: 'absolute',
          bottom: 4,
          right: 8,
          textAlign: 'right',
        }}
      >
        <TruncatedText variant="caption" sx={{ fontWeight: 300 }}>
          {numWorkflowsUsingMsg}
        </TruncatedText>
      </Box>
    </Box>
  );
};
