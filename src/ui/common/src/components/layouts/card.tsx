import Box from '@mui/material/Box';
import { styled } from '@mui/material/styles';

import { theme } from '../../styles/theme/theme';

export const CardPadding = '16px';

export const Card = styled(Box)(() => {
  return {
    borderRadius: 4,
    '&:hover': {
      backgroundColor: theme.palette.blue[50],
    },
    minWidth: '450px',
    padding: CardPadding,
  };
});
