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

// Maximum number of rows to show before creating a new column.
const maxNumRows = 3;

function chunkList(list: string[], chunkSize: number): string[][] {
  return Array.from({ length: Math.ceil(list.length / chunkSize) }, (_, index) =>
      list.slice(index * chunkSize, index * chunkSize + chunkSize)
  );
}

// The format is "Label: Value". The label width is set to the maximum of all the provided labels.
export const ResourceCardText: React.FC<ResourceCardTextProps> = ({
  labels,
  values,
}) => {
  // Wrapper around useTextWidth to find the label length + padding in pixels.
  const useLabelWidth = (label: string): number => {
    return useTextWidth({ text: label }) + paddingBetweenLabelAndValue;
  };
  const labelWidthNum = Math.max(...labels.map(useLabelWidth));

  // Chunk the fields into their respective columns.
  const chunkedLabels = chunkList(labels, maxNumRows)
  const chunkedValues = chunkList(values, maxNumRows)

  console.log("Labels: ", chunkedLabels)

  return (
    <Box sx={{ display: 'flex', flexDirection: 'row'}}>
      {chunkedLabels.map((labelsForColumn, colIndex) => (
          <Box sx={{ display: 'flex', flexDirection: 'column' }}>
            {labelsForColumn.map((_, index) => (
                <Box key={index} sx={{ display: 'flex', flexWrap: 'nowrap' }}>
                  <TruncatedText
                      variant="body2"
                      sx={{ fontWeight: 300, width: `${labelWidthNum}px` }}
                  >
                    {labelsForColumn[index]}
                  </TruncatedText>
                  <TruncatedText
                      variant="body2"
                      sx={{ width: `calc(100% - ${labelWidthNum}px)` }}
                  >
                    {chunkedValues[colIndex][index]}
                  </TruncatedText>
                </Box>
            ))}
          </Box>
      ))}
    </Box>
  );


  // return (
  //     <Box sx={{ display: 'flex', flexDirection: 'column' }}>
  //       {labels.map((_, index) => (
  //           <Box key={index} sx={{ display: 'flex', flexWrap: 'nowrap' }}>
  //             <TruncatedText
  //                 variant="body2"
  //                 sx={{ fontWeight: 300, width: `${labelWidthNum}px` }}
  //             >
  //               {labels[index]}
  //             </TruncatedText>
  //             <TruncatedText
  //                 variant="body2"
  //                 sx={{ width: `calc(100% - ${labelWidthNum}px)` }}
  //             >
  //               {values[index]}
  //             </TruncatedText>
  //           </Box>
  //       ))}
  //     </Box>
  // );
};
