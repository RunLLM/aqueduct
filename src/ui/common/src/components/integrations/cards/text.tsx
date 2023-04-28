import { Box, Typography } from '@mui/material';
import { styled } from '@mui/material/styles';
import { useTextWidth } from '@tag0/use-text-width';
import React from 'react';

// Use in place of Typography when you need the text to be truncated with an ellipsis.
export const TruncatedText = styled(Typography)(() => {
  return {
    whiteSpace: 'nowrap',
    textOverflow: 'ellipsis',
    overflow: 'hidden',
  };
});

type ResourceCardTextProps = {
  labels: string[];
  values: string[];
};

const paddingBetweenLabelAndValue = 8; // in pixels

// The format is "Label: Value". The label width is set to the maximum of all the provided labels.
// The maximum number of fields per
export const ResourceCardText: React.FC<ResourceCardTextProps> = ({
  labels,
  values,
}) => {
  // Wrapper around useTextWidth to find the label length + padding in pixels.
  const useLabelWidth = (label: string): number => {
    return useTextWidth({ text: label }) + paddingBetweenLabelAndValue;
  };
  const labelWidthNum = Math.max(...labels.map(useLabelWidth));

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      {labels.map((_, index) => (
        <Box key={index} sx={{ display: 'flex', flexWrap: 'nowrap' }}>
          <TruncatedText
            variant="body2"
            sx={{ fontWeight: 300, width: `${labelWidthNum}px` }}
          >
            {labels[index]}
          </TruncatedText>
          <TruncatedText
            variant="body2"
            sx={{ width: `calc(100% - ${labelWidthNum}px)` }}
          >
            {values[index]}
          </TruncatedText>
        </Box>
      ))}
    </Box>
  );
};
