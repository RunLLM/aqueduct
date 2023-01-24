import { Box, Typography } from '@mui/material';
import { ComponentMeta, ComponentStory } from '@storybook/react';
import React from 'react';

import { StatusIndicator } from '../components/workflows/workflowStatus';
import ExecutionStatus from '../utils/shared';

interface StatusIndicatorStoryProps {
  /**
   * Execution status to render
   */
  status: ExecutionStatus;
  /**
   * Label for execution status
   */
  label: string;
}

export default {
  title: 'Components/Status Indicator',
  component: StatusIndicator,
  argTypes: {},
} as ComponentMeta<typeof StatusIndicator>;

const Template: ComponentStory<typeof StatusIndicator> = (
  args: StatusIndicatorStoryProps
) => (
  <Box display="flex" alignItems="center">
    <StatusIndicator {...args} />
    <Typography variant="body1" sx={{ marginLeft: '8px' }}>
      {args.label}
    </Typography>
  </Box>
);

export const LargeIndicator = Template.bind({});
LargeIndicator.args = {
  status: ExecutionStatus.Succeeded,
  label: 'Succeeded',
  size: '50px',
};

export const BlackIndicator = Template.bind({});
BlackIndicator.args = {
  status: ExecutionStatus.Succeeded,
  label: 'Succeeded',
  monochrome: 'black',
};

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
