//import styled from '@emotion/styled';
import { styled } from '@mui/material/styles';
import IconButton, { iconButtonClasses } from '@mui/material/IconButton';

const AqueductIconButton = styled(IconButton)({
  [`&.${iconButtonClasses.root}`]: {
    '&:hover': {
      backgroundColor: 'transparent',
    },
  },
});

AqueductIconButton.defaultProps = {
  disableRipple: true,
};

export { AqueductIconButton as IconButton };
