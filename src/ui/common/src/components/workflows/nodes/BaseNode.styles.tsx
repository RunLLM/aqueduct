import Box from '@mui/material/Box';
import { styled } from '@mui/material/styles';

const BaseNode = styled(Box)({
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
  justifyContent: 'center',
  borderRadius: '8px',
  borderStyle: 'solid',
  borderWidth: '2px',
  width: '310px',
  height: '160px',
  maxWidth: '310px',
  maxHeight: '160px',
});

export { BaseNode };
