//import styled from '@emotion/styled';
import IconButton, { iconButtonClasses } from '@mui/material/IconButton';
import { styled } from '@mui/material/styles';

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
