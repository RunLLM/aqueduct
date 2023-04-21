import { Box, Typography } from '@mui/material';
import { styled } from '@mui/material/styles';
import React from 'react';

// Use in place of Typography when you need the text to be truncated with an ellipsis.
export const TruncatedText = styled(Typography)(() => {
  return {
    fontFamily: 'Roboto',
    whiteSpace: 'nowrap',
    textOverflow: 'ellipsis',
    overflow: 'hidden',
  };
});

type CardTextProps = {
  category: string;
  value: string;
  categoryWidth: string;
};

// Use when filling in a single line of text within an integration card.
// The format is "Category: Value". The category width should be set to the minimum
// width required to display the longest category name.
export const CardTextEntry: React.FC<CardTextProps> = ({
  category,
  value,
  categoryWidth,
}) => {
  return (
    <Box sx={{ display: 'flex', flexWrap: 'nowrap' }}>
      <TruncatedText
        variant="body2"
        sx={{ fontWeight: 300, width: categoryWidth }}
      >
        {category}
      </TruncatedText>
      <TruncatedText
        variant="body2"
        sx={{ width: `calc(100% - ${categoryWidth})` }}
      >
        {value}
      </TruncatedText>
    </Box>
  );
};
