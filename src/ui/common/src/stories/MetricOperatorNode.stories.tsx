import { ComponentMeta, ComponentStory } from '@storybook/react';
import React from 'react';

import MetricOperatorNode from '../components/workflows/nodes/MetricOperatorNode';
import { ReactFlowNodeData, ReactflowNodeType } from '../utils/reactflow';

export default {
  title: 'Components/Dag Node',
  component: MetricOperatorNode,
  argTypes: {},
} as ComponentMeta<typeof MetricOperatorNode>;

// TODO: Add a ReactFlowCanvas here and wrap the MetricOperatorNode in the canvas.
// After we get this, we'll have an easy way to iterate on nodes.
const Template: ComponentStory<typeof MetricOperatorNode> = (args) => (
  <MetricOperatorNode {...args} />
);

export const MetricOperatorNodeStory = Template.bind({});
const mockData: ReactFlowNodeData = {
  nodeId: 'metricOperatorNodeStory',
  nodeType: ReactflowNodeType.Operator,
  onChange: () => {
    console.log('mock onChange');
  },
  onConnect: () => {
    console.log('mock onConnect');
  },
  label: 'metricOperatorStory',
  result: '1.234',
};

MetricOperatorNodeStory.args = {
  data: mockData,
  isConnectable: false,
};
