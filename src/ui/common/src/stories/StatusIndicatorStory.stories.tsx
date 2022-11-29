import { Box, Typography } from '@mui/material';
import { ComponentMeta, ComponentStory } from '@storybook/react';
import React from 'react';

import { StatusIndicator } from '../components/workflows/workflowStatus';
import ExecutionStatus from '../utils/shared';

interface StatusIndicatorStoryProps {
  status: ExecutionStatus;
  label: string;
}

export default {
  title: 'Status Indicator',
  component: StatusIndicator,
  // More on argTypes: https://storybook.js.org/docs/react/api/argtypes
  argTypes: {},
} as ComponentMeta<typeof StatusIndicator>;

// More on component templates: https://storybook.js.org/docs/react/writing-stories/introduction#using-args
const Template: ComponentStory<typeof StatusIndicator> = (
  args: StatusIndicatorStoryProps
) => (
  <Box display="flex">
    <StatusIndicator {...args} />
    <Typography variant="body1" sx={{ marginLeft: '8px' }}>
      {args.label}
    </Typography>
  </Box>
);

export const CanceledStatusIndicator = Template.bind({});
CanceledStatusIndicator.args = {
  status: ExecutionStatus.Canceled,
  label: 'Canceled',
};

export const FailedStatusIndicator = Template.bind({});
FailedStatusIndicator.args = {
  status: ExecutionStatus.Failed,
  label: 'Failed',
};

export const PendingStatusIndicator = Template.bind({});
PendingStatusIndicator.args = {
  status: ExecutionStatus.Pending,
  label: 'Pending',
};

export const RegisteredStatusIndicator = Template.bind({});
RegisteredStatusIndicator.args = {
  status: ExecutionStatus.Registered,
  label: 'Registered',
};

export const RunningStatusIndicator = Template.bind({});
RunningStatusIndicator.args = {
  status: ExecutionStatus.Running,
  label: 'Running',
};

export const SucceededStatusIndicator = Template.bind({});
SucceededStatusIndicator.args = {
  status: ExecutionStatus.Succeeded,
  label: 'Succeeded',
};

export const UnknownStatusIndicator = Template.bind({});
UnknownStatusIndicator.args = {
  status: ExecutionStatus.Unknown,
  label: 'Unknown',
};
