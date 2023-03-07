import Box from '@mui/material/Box';
import { styled } from '@mui/material/styles';

const BaseNode = styled(Box)({
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
  justifyContent: 'center',
  borderRadius: '8px',
  borderStyle: 'solid',
  borderWidth: '3px',
  width: '400px',
  height: '120px',
  maxWidth: '400px',
  maxHeight: '120px',
  textOverflow: 'ellipsis',
});

export { BaseNode };
