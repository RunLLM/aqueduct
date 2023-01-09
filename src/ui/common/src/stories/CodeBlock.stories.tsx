import { ComponentMeta, ComponentStory } from '@storybook/react';
import React from 'react';
import CodeBlock from '../components/CodeBlock';

// More on default export: https://storybook.js.org/docs/react/writing-stories/introduction#default-export
export default {
    title: 'Components/Code Block',
    component: CodeBlock,
    // More on argTypes: https://storybook.js.org/docs/react/api/argtypes
    argTypes: {
    },
} as ComponentMeta<typeof CodeBlock>;

// More on component templates: https://storybook.js.org/docs/react/writing-stories/introduction#using-args
const Template: ComponentStory<typeof CodeBlock> = (args) => <CodeBlock {...args} />;

const apiConnectionSnippet = `import aqueduct
client = aqueduct.Client(
    "QYIF0375KE4Z2GXSB6T8RNJLVMA9WCPH",
    "http://localhost:8080"
)`;

const sqlSnippet = "SELECT * FROM CUSTOMERS;"

export const PythonExample = Template.bind({});
// More on args: https://storybook.js.org/docs/react/writing-stories/args
PythonExample.args = {
    language: "python",
    children: apiConnectionSnippet
};

export const SQLExample = Template.bind({});
SQLExample.args = {
    language: "sql",
    children: sqlSnippet
};
