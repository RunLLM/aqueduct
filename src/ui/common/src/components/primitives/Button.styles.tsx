import Button, { buttonClasses } from '@mui/material/Button';
import { styled } from '@mui/material/styles';

import { theme } from '../../styles/theme/theme';

const AqueductButton = styled(Button)(() => {
  return {
    [`&.${buttonClasses.root}`]: {
      textTransform: 'none',
      boxShadow: 'none',
      fontSize: '16px',
      disableElevation: true,

      // Theming for primary colored buttons.
      [`&.${buttonClasses.containedPrimary}`]: {
        color: 'white',
        backgroundColor: theme.palette.blue[900],
        '&:hover': {
          backgroundColor: theme.palette.blue[700],
        },
        [`&.${buttonClasses.disabled}`]: {
          backgroundColor: theme.palette.gray[700],
        },
      },

      [`&.${buttonClasses.outlinedPrimary}`]: {
        color: theme.palette.blue[900],
        borderColor: theme.palette.blue[900],
      },

      // Theming for secondary colored buttons.
      [`&.${buttonClasses.containedSecondary}`]: {
        color: theme.palette.darkGray,
        backgroundColor: theme.palette.gray[200],
        '&:hover': {
          backgroundColor: theme.palette.gray[500],
        },
        [`&.${buttonClasses.disabled}`]: {
          color: theme.palette.gray[100],
          backgroundColor: theme.palette.gray[300],
        },
      },
      [`&.${buttonClasses.outlinedSecondary}`]: {
        color: theme.palette.darkGray,
        borderColor: theme.palette.gray[700],
      },
    },
  };
});

AqueductButton.defaultProps = {
  disableRipple: true,
  variant: 'contained',
};

export { AqueductButton as Button };
