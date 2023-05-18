import { Box, Typography } from '@mui/material';
import TextField from '@mui/material/TextField';
import React, { ChangeEvent } from 'react';
import { useFormContext } from 'react-hook-form';

type IntegrationTextFieldProps = {
  label: string;
  description: string;
  warning?: string;
  spellCheck: boolean;
  required: boolean;
  placeholder?: string;
  onChange?: (
    event: ChangeEvent<HTMLTextAreaElement | HTMLInputElement>
  ) => void;
  type?: string;
  disabled?: boolean;
  disableReason?: string;
  autoComplete?: string;
  name: string;
};

export const IntegrationTextInputField: React.FC<IntegrationTextFieldProps> = ({
  label,
  description,
  warning,
  spellCheck,
  required,
  placeholder,
  onChange,
  type,
  disabled,
  disableReason,
  autoComplete,
  name,
}) => {
  const { register } = useFormContext();

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
          name={name}
          spellCheck={spellCheck}
          required={required}
          placeholder={placeholder}
          type={type ? type : null}
          fullWidth={true}
          size={'small'}
          disabled={disabled}
          helperText={disabled ? disableReason : undefined}
          autoComplete={autoComplete}
          {...register(name, { required })}
          onChange={onChange}
        />
      </Box>
    </Box>
  );
};
