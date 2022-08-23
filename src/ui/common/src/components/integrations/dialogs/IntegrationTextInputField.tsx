import { Box, Typography } from '@mui/material';
import TextField from '@mui/material/TextField';
import React, { ChangeEvent } from 'react';

type IntegrationTextFieldProps = {
  label: string;
  description: string;
  warning?: string;
  spellCheck: boolean;
  required: boolean;
  placeholder?: string;
  onChange: (
    event: ChangeEvent<HTMLTextAreaElement | HTMLInputElement>
  ) => void;
  value: string;
  type?: string;
  disabled?: boolean;
  disableReason?: string;
};

export const IntegrationTextInputField: React.FC<IntegrationTextFieldProps> = ({
  label,
  description,
  warning,
  spellCheck,
  required,
  placeholder,
  onChange,
  value,
  type,
  disabled,
  disableReason,
}) => {
  return (
    <Box sx={{ mt: 2 }}>
      <Box sx={{ my: 1 }}>
        <Typography variant="body1" sx={{ fontWeight: 'bold' }}>
          {label}
        </Typography>
        <Typography variant="body2" sx={{ color: 'darkGray' }}>
          {description}
          <em>{warning}</em>
        </Typography>
      </Box>
      <Box>
        <TextField
          spellCheck={spellCheck}
          required={required}
          placeholder={placeholder}
          onChange={onChange}
          value={value}
          type={type ? type : null}
          fullWidth={true}
          size={'small'}
          disabled={disabled}
          helperText={disabled ? disableReason : undefined}
        />
      </Box>
    </Box>
  );
};
