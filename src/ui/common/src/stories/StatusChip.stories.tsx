import { ComponentMeta, ComponentStory } from '@storybook/react';
import React from 'react';

import { StatusChip } from '../components/workflows/workflowStatus';
import ExecutionStatus from '../utils/shared';

export default {
  title: 'Components/Status Chip',
  component: StatusChip,
  argTypes: {},
} as ComponentMeta<typeof StatusChip>;

const Template: ComponentStory<typeof StatusChip> = (args) => (
  <StatusChip {...args} />
);

export const CanceledStatusChip = Template.bind({});
CanceledStatusChip.args = {
  status: ExecutionStatus.Canceled,
};

export const FailedStatusChip = Template.bind({});
FailedStatusChip.args = {
  status: ExecutionStatus.Failed,
};

export const PendingStatusChip = Template.bind({});
PendingStatusChip.args = {
  status: ExecutionStatus.Pending,
};

export const RegisteredStatusChip = Template.bind({});
RegisteredStatusChip.args = {
  status: ExecutionStatus.Registered,
};

export const RunningStatusChip = Template.bind({});
RunningStatusChip.args = {
  status: ExecutionStatus.Running,
};

export const SucceededStatusChip = Template.bind({});
SucceededStatusChip.args = {
  status: ExecutionStatus.Succeeded,
};

export const UnknownStatusChip = Template.bind({});
UnknownStatusChip.args = {
  status: ExecutionStatus.Unknown,
};
