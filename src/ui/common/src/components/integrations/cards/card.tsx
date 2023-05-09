import Box from '@mui/material/Box';
import React from 'react';

import { Integration } from '../../../utils/integrations';
import ExecutionStatus from '../../../utils/shared';
import { StatusIndicator } from '../../workflows/workflowStatus';
import IntegrationLogo from '../logo';
import { ResourceFieldsDetailsCard } from './resourceFieldsDetailsCard';
import { TruncatedText } from './text';

type IntegrationProps = {
  integration: Integration;

  // Eg: "Used by 2 workflows"
  numWorkflowsUsingMsg: string;
};

export const IntegrationCard: React.FC<IntegrationProps> = ({
  integration,
  numWorkflowsUsingMsg,
}) => {
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Box sx={{ display: 'flex', flexDirection: 'row', alignItems: 'center' }}>
        {/*If the execution state doesn't exist, we assume the integration succeeded.*/}
        <StatusIndicator
          status={integration.exec_state?.status || ExecutionStatus.Succeeded}
          size="16px"
        />

        {/* Subtract the width of the status indicator, padding, and logo respectively. */}
        <Box
          sx={{ mx: 1, flex: 1, maxWidth: `calc(100% - 16px - 16px - 24px)` }}
        >
          <TruncatedText sx={{ fontWeight: 400 }} variant="h6">
            {integration.name}
          </TruncatedText>
        </Box>
        <IntegrationLogo service={integration.service} size="small" activated />
      </Box>

      {/*Leave this empty if integration.createdAt isn't set.*/}
      <TruncatedText
        variant="caption"
        marginBottom={1}
        sx={{ fontWeight: 300 }}
      >
        {integration.createdAt
          ? new Date(integration.createdAt * 1000).toLocaleString()
          : '  '}
      </TruncatedText>

      <ResourceFieldsDetailsCard
        integration={integration}
        detailedView={false}
      />

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
