import Box from '@mui/material/Box';
import { styled } from '@mui/material/styles';

import { theme } from '../../styles/theme/theme';

export const CardPadding = '8px';

export const Card = styled(Box)(() => {
  return {
    borderRadius: 4,
    '&:hover': {
      backgroundColor: theme.palette.blue[50],
    },
    width: '325px',
    height: '156px',
    padding: CardPadding,
  };
});
