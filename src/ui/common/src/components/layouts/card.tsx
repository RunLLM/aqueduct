import Box from '@mui/material/Box';
import { styled } from '@mui/material/styles';

import { theme } from '../../styles/theme/theme';

export const CardPadding = '8px';

export const Card = styled(Box)(() => {
  return {
    borderRadius: 4,
    '&:hover': {
      backgroundColor: theme.palette.gray[250],
    },
    backgroundColor: theme.palette.gray[25],
    width: '325px',
    height: '150px',
    padding: CardPadding,
    position: 'relative',
  };
});
