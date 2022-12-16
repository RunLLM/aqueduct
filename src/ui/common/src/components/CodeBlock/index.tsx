import React from 'react';
import SyntaxHighlighter from 'react-syntax-highlighter';
import { docco } from 'react-syntax-highlighter/dist/cjs/styles/hljs';

type Props = {
  /**
   * Programming language to be syntax highlighted in the code block.
   */
  language: string;
  /**
   * Code string to be rendered in the code block.
   */
  children: string;
};

/**
 * Component used to show syntax highlighted code on the UI.
 */
export const CodeBlock: React.FC<Props> = ({ language, children }) => {
  return (
    <SyntaxHighlighter
      language={language}
      style={docco}
      customStyle={{
        borderRadius: 4,
        padding: '15px',
        // overrides built-in margin
        margin: '0px',
      }}
    >
      {children}
    </SyntaxHighlighter>
  );
};

export default CodeBlock;
