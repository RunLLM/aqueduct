import { ComponentMeta } from '@storybook/react';
import React from 'react';

import { ResourceHeaderDetailsCard } from '../components/resources/cards/headerDetailsCard';
import { Resource, SlackConfig } from '../utils/resources';
import ExecutionStatus from '../utils/shared';

export const ResourceHeaderDetailsCardStory: React.FC = () => {
  const testResource: Resource = {
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
      resource={testResource}
      numWorkflowsUsingMsg="Not currently in use"
    />
  );
};

export default {
  title: 'Components/Resource Header Details Card',
  component: ResourceHeaderDetailsCard,
  argTypes: {},
} as ComponentMeta<typeof ResourceHeaderDetailsCard>;
