import React from 'react';
import SyntaxHighlighter from 'react-syntax-highlighter';
import { docco } from 'react-syntax-highlighter/dist/cjs/styles/hljs';

type Props = {
  language: string;
  children: string;
};

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
