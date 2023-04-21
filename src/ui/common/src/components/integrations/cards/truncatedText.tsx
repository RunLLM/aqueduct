import {Typography} from "@mui/material";
import {styled} from "@mui/material/styles";

// Use in place of Typography when you need the text to be truncated with an ellipsis.
export const TruncatedText = styled(Typography)(() => {
  return {
    whiteSpace: 'nowrap',
    textOverflow: 'ellipsis',
    overflow: 'hidden',
  };
});