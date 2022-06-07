import Box from '@mui/material/Box';
import { styled } from '@mui/material/styles';

import { theme } from '../../styles/theme/theme';

export const Card = styled(Box)(({ theme: Theme }) => {
  return {
    borderRadius: 4,
    '&:hover': {
      backgroundColor: theme.palette.blue[50],
    },
    minWidth: '450px',
    padding: '16px',
  };
});
