import LoadingButton, { loadingButtonClasses } from '@mui/lab/LoadingButton';
import { buttonClasses } from '@mui/material/Button';
import { styled } from '@mui/material/styles';

import { theme } from '../../styles/theme/theme';

const AqueductLoadingButton = styled(LoadingButton)(() => {
  return {
    [`&.${loadingButtonClasses.root}`]: {
      textTransform: 'none',
      boxShadow: 'none',
      fontSize: '16px',
      disableElevation: true,

      // Theming for primary colored buttons.
      [`&.${buttonClasses.containedPrimary}`]: {
        color: 'white',
        backgroundColor: theme.palette.blue[900],
        [`&.${buttonClasses.disabled}`]: {
          color: 'white',
          backgroundColor: theme.palette.gray[700],
        },
        '&:hover': {
          backgroundColor: theme.palette.blue[700],
        },
      },
    },
  };
});

export { AqueductLoadingButton as LoadingButton };
