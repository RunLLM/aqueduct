import { Box, Checkbox, Typography } from '@mui/material';
import React from 'react';

import { theme } from '../../styles/theme/theme';

type Props = {
  checked: boolean;
  disabled?: boolean;
  onChange: (checked: boolean) => void;
  children?: string | JSX.Element | (string | JSX.Element)[];
};

const CheckboxEntry: React.FC<Props> = ({
  checked,
  disabled,
  onChange,
  children,
}) => {
  return (
    <Box display="flex" flexDirection="row" alignContent="center">
      <Checkbox
        checked={checked}
        disabled={disabled}
        onChange={(event) => onChange(event.target.checked)}
        // Overrides default checkbox behavior.
        // The `padding` is particularly important as the checkbox has a default padding of 9.
        sx={{
          padding: 0,
          '&.Mui-checked': { color: theme.palette.blue[700] },
          '&.Mui-disabled': { color: theme.palette.gray[700] },
        }}
      />
      <Typography
        variant="body1"
        color={disabled ? 'gray.700' : 'black'}
        marginLeft={1}
      >
        {children}
      </Typography>
    </Box>
  );
};

export default CheckboxEntry;
