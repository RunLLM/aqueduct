import { Typography } from '@mui/material';
import { ComponentMeta, ComponentStory } from '@storybook/react';
import React from 'react';

// More on default export: https://storybook.js.org/docs/react/writing-stories/introduction#default-export
export default {
  title: 'Primitives/Typography',
  component: Typography,
  // More on argTypes: https://storybook.js.org/docs/react/api/argtypes
  argTypes: {},
} as ComponentMeta<typeof Typography>;

// More on component templates: https://storybook.js.org/docs/react/writing-stories/introduction#using-args
const Template: ComponentStory<typeof Typography> = (args) => (
  <Typography {...args} />
);

const sampleText = 'The quick brown fox jumps over the lazy dog.';

// More on args: https://storybook.js.org/docs/react/writing-stories/args
export const HeadingOne = Template.bind({});
HeadingOne.args = {
  variant: 'h1',
  children: sampleText,
};

export const HeadingTwo = Template.bind({});
HeadingTwo.args = {
  variant: 'h2',
  children: sampleText,
};

export const HeadingThree = Template.bind({});
HeadingThree.args = {
  variant: 'h3',
  children: sampleText,
};

export const HeadingFour = Template.bind({});
HeadingFour.args = {
  variant: 'h4',
  children: sampleText,
};

export const HeadingFive = Template.bind({});
HeadingFive.args = {
  variant: 'h5',
  children: sampleText,
};

export const HeadingSix = Template.bind({});
HeadingSix.args = {
  variant: 'h6',
  children: sampleText,
};

export const BodyOne = Template.bind({});
BodyOne.args = {
  variant: 'body1',
  children: sampleText,
};

export const BodyTwo = Template.bind({});
BodyTwo.args = {
  variant: 'body2',
  children: sampleText,
};

export const ButtonText = Template.bind({});
ButtonText.args = {
  variant: 'button',
  children: sampleText,
};

export const CaptionText = Template.bind({});
CaptionText.args = {
  variant: 'caption',
  children: sampleText,
};

export const OverlineText = Template.bind({});
OverlineText.args = {
  variant: 'overline',
  children: sampleText,
};
