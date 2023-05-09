import Box from '@mui/material/Box';
import React from 'react';

import {AqueductComputeConfig, Integration} from '../../../utils/integrations';
import ExecutionStatus, {ExecState} from '../../../utils/shared';
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

  // If the exec_state doesn't exist on the resource, we assume the integration succeeded.
  // Moreover, if the resource is a Aqueduct compute, we need to also check the status of any
  // registered conda resouce!
  let status = integration.exec_state?.status || ExecutionStatus.Succeeded;
  if (integration.service == "Aqueduct" && integration.exec_state.status == ExecutionStatus.Succeeded){
    const aqConfig = integration.config as AqueductComputeConfig
    if (aqConfig.conda_config_serialized) {
      const serialized_conda_exec_state = JSON.parse(aqConfig.conda_config_serialized)["exec_state"]
      const conda_exec_state = JSON.parse(serialized_conda_exec_state) as ExecState
      status = conda_exec_state.status
    }
  }

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Box sx={{ display: 'flex', flexDirection: 'row', alignItems: 'center' }}>
        <StatusIndicator
          status={status}
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
