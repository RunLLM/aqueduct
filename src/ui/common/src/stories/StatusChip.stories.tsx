import { ComponentMeta, ComponentStory } from '@storybook/react';
import React from 'react';

import { StatusChip } from '../components/workflows/workflowStatus';
import ExecutionStatus from '../utils/shared';

export default {
  title: 'Status Chip',
  component: StatusChip,
  // More on argTypes: https://storybook.js.org/docs/react/api/argtypes
  argTypes: {
    //backgroundColor: { control: 'color' },
  },
} as ComponentMeta<typeof StatusChip>;

// More on component templates: https://storybook.js.org/docs/react/writing-stories/introduction#using-args
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
