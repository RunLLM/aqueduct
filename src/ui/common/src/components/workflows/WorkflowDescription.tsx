import { Typography } from '@mui/material';
import React from 'react';
import Markdown from 'react-markdown';
import { visitParents } from 'unist-util-visit-parents';

import { WorkflowResponse } from '../../handlers/responses/Workflow';
import style from '../../styles/markdown.module.css';

type Props = {
  workflow: WorkflowResponse;
};

const WorkflowDescription: React.FC<Props> = ({ workflow }) => {
  /**
   * Wrap text in a `custom-typography` tag
   */
  function rehypeWrapText() {
    return function wrapTextTransform(tree) {
      visitParents(tree, 'text', (node, ancestors) => {
        if (ancestors[ancestors.length - 1]?.tagName !== 'custom-typography') {
          node.type = 'element';
          node.tagName = 'custom-typography';
          node.children = [{ type: 'text', value: node.value }];
        }
      });
    };
  }

  return (
    <Markdown
      className={style.reactMarkdown}
      rehypePlugins={[rehypeWrapText]}
      components={{
        'custom-typography': ({ children }) => (
          <Typography variant="body1">{children}</Typography>
        ),
      }}
    >
      {workflow.description === '' ? '*No description.*' : workflow.description}
    </Markdown>
  );
};

export default WorkflowDescription;
