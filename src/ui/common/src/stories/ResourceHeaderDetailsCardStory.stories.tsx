import React from 'react';

import { ResourceHeaderDetailsCard } from '../components/integrations/cards/headerDetailsCard';
import { Integration, SlackConfig } from '../utils/integrations';
import ExecutionStatus from '../utils/shared';

export const ResourceHeaderDetailsCardStory: React.FC = () => {
  const testIntegration: Integration = {
    id: '20',
    service: 'Slack',
    name: 'Another Slack Longer Name',
    config: {
      token: 'xoxb-123456789012-1234567890123-123456789012345678901234',
      channels_serialized: '["#general"]',
      level: 'warning',
      enabled: 'true',
    } as SlackConfig,
    createdAt: Date.now() / 1000,
    exec_state: {
      status: ExecutionStatus.Succeeded,
    },
  };

  return (
    <ResourceHeaderDetailsCard
      integration={testIntegration}
      numWorkflowsUsingMsg="Not currently in use"
    />
  );
};

export default ResourceHeaderDetailsCardStory;
