import React from 'react';
import SyntaxHighlighter from 'react-syntax-highlighter';
import { docco } from 'react-syntax-highlighter/dist/cjs/styles/hljs';

type Props = {
  language: string;
  children: any;
};

export const CodeBlock: React.FC<Props> = ({ language, children }) => {
  console.log('codeblock children: ', children);
  return (
    <SyntaxHighlighter
      language={language}
      style={docco}
      customStyle={{ borderRadius: 4, padding: '15px' }}
    >
      {children}
    </SyntaxHighlighter>
  );
};

export default CodeBlock;
