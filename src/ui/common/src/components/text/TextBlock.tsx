import Typography from '@mui/material/Typography';
import React from 'react';

type Props = {
  text: string;
};

const TextBlock: React.FC<Props> = ({ text }) => {
  return (
    <Typography sx={{ fontFamily: 'Monospace', whiteSpace: 'pre-wrap' }}>
      {text}
    </Typography>
  );
};

export default TextBlock;
