import { ComponentMeta, ComponentStory } from '@storybook/react';
import React from 'react';

import { InfoTooltip } from '../components/pages/components/InfoTooltip';

export default {
  title: 'Example/InfoTooltip',
  component: InfoTooltip,
  parameters: {
    // More on Story layout: https://storybook.js.org/docs/react/configure/story-layout
    layout: 'fullscreen',
  },
} as ComponentMeta<typeof InfoTooltip>;

const Template: ComponentStory<typeof InfoTooltip> = (args) => (
  <InfoTooltip {...args} />
);

export const RightTooltip = Template.bind({});
RightTooltip.args = {
  tooltipText:
    'This is the tooltip content. It appears to the right of the button.',
  placement: 'right',
};

export const BottomTooltip = Template.bind({});
RightTooltip.args = {
  tooltipText:
    'This is the tooltip content. It appears to the bottom of the button.',
  placement: 'bottom',
};
