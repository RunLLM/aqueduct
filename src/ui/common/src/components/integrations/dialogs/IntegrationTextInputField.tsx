import React, {ChangeEvent} from "react";
import {Box, Typography} from "@mui/material";
import TextField from "@mui/material/TextField";

type IntegrationTextFieldProps = {
    label: string;
    description: string;
    spellCheck: boolean;
    required: boolean;
    placeholder: string;
    onChange: (
        event: ChangeEvent<HTMLTextAreaElement | HTMLInputElement>
    ) => void;
    value: string;
    type?: string;
    disabled?: boolean;
};

export const IntegrationTextInputField: React.FC<IntegrationTextFieldProps> = ({
   label,
   description,
   spellCheck,
   required,
   placeholder,
   onChange,
   value,
   type,
   disabled,
   }) => {
    return (
        <Box sx={{ mt: 2 }}>
            <Box sx={{ my: 1 }}>
                <Typography variant="body1" sx={{ fontWeight: 'bold' }}>
                    {label}
                </Typography>
                <Typography variant="body2" sx={{ color: 'darkGray' }}>
                    {description}
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
                />
            </Box>
        </Box>
    );
};