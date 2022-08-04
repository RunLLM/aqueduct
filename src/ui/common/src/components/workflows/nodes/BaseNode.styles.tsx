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
  padding: '10px',
  maxWidth: '250px',
  minHeight: '140px',
  maxHeight: '250px',
});

export { BaseNode };
