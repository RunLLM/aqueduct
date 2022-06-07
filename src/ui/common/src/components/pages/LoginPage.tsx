import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import React, { ChangeEvent, useState } from 'react';
import cookies from 'react-cookies';
import { useNavigate } from 'react-router-dom';

import fetchUser from '../../utils/fetchUser';
import setUser from '../hooks/setUser';

export const LoginPage: React.FC = () => {
  const [validationError, setValidationError] = useState<boolean>(false);
  const [errorMsg, setErrorMsg] = useState<string>('');
  const [apiKey, setApiKey] = useState<string>(
    cookies.load('aqueduct-api-key')
  );
  const navigate = useNavigate();

  const onApiKeyTextFieldChanged = (
    event: ChangeEvent<HTMLTextAreaElement | HTMLInputElement>
  ) => {
    const input = event.target.value;
    if (input.length > 0) {
      setValidationError(false);
      setErrorMsg('');
    } else {
      setValidationError(true);
      setErrorMsg('Api key should not be empty.');
    }
    setApiKey(input);
  };

  const onGetStartedClicked = async (event: React.MouseEvent) => {
    event.preventDefault();
    const { success } = await fetchUser(apiKey);
    if (!apiKey || apiKey.length === 0) {
      setValidationError(true);
      setErrorMsg('Api key should not be empty.');
    } else if (!success) {
      setValidationError(true);
      setErrorMsg(
        "Invalid key, please copy the key from 'aqueduct apikey' outputs."
      );
    } else {
      setValidationError(false);
      setUser(apiKey);
      navigate('/');
    }
  };

  return (
    <Box
      sx={{
        width: '100%',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
      }}
    >
      <Box sx={{ width: '350px' }}>
        <Box
          marginTop="175px"
          sx={{
            width: '100%',
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            flexDirection: 'column',
          }}
        >
          <img
            src={
              'https://aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com/webapp/logos/aqueduct_logo_color_on_white.png'
            }
            width="150px"
            height="150px"
          />
          <TextField
            error={validationError}
            sx={{ marginTop: 2 }}
            fullWidth={true}
            size="small"
            id="outlined-basic"
            label={'Please enter an API Key'}
            helperText={errorMsg}
            variant="outlined"
            onChange={onApiKeyTextFieldChanged}
          />
        </Box>
        <Button
          onClick={onGetStartedClicked}
          sx={{ marginTop: 1 }}
          fullWidth={true}
          color="primary"
          variant="contained"
        >
          Get Started
        </Button>
      </Box>
    </Box>
  );
};

export default LoginPage;
