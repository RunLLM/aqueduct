import Box from '@mui/material/Box';
import React from 'react';

type Props = {
  children: JSX.Element[];
};

const ButtonGroup: React.FC<Props> = ({ children }) => {
  return (
    <Box
      display="flex"
      flexDirection="row"
      alignContent="center"
      alignItems="center"
      sx={{
        paddingX: '2px',
        margin: '2px',
        height: 'fit-content',
      }}
    >
      {children}
    </Box>
  );
};

export default ButtonGroup;
