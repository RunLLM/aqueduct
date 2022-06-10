import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';
import React, { ChangeEvent, useEffect, useState } from 'react';
import { useCookies } from 'react-cookie';

import fetchUser from '../../utils/fetchUser';
import { Button } from '../primitives/Button.styles';

export const LoginPage: React.FC = () => {
  useEffect(() => {
    document.title = 'Login | Aqueduct';
  }, []);

  const [cookies, setCookie, removeCookie] = useCookies(['aqueduct-api-key']);

  const [isAuthed, setIsAuthed] = useState<boolean>(false);
  const [validationError, setValidationError] = useState<boolean>(false);
  const [errorMsg, setErrorMsg] = useState<string>('');
  const [apiKey, setApiKey] = useState<string>(cookies['aqueduct-api-key']);

  const onApiKeyTextFieldChanged = (
    event: ChangeEvent<HTMLTextAreaElement | HTMLInputElement>
  ) => {
    const input = event.target.value;
    if (input.length > 0) {
      setValidationError(false);
      setErrorMsg('');
    } else {
      setValidationError(true);
      setErrorMsg('API Key should not be empty.');
    }
    setApiKey(input);
  };

  const onGetStartedClicked = async (event: React.MouseEvent) => {
    event.preventDefault();
    const { success } = await fetchUser(apiKey);

    if (!success) {
      setValidationError(true);
      setErrorMsg(
        'Invalid API Key. You can find your API Key by running `aqueduct apikey` on the machine where Aqueduct is running.'
      );
    } else {
      setCookie('aqueduct-api-key', apiKey, { path: '/' });
      await new Promise((r) => setTimeout(r, 100));
      setValidationError(false);

      // Once we validate, we force a reload of the page. This is because React
      // doesn't give us an easy way to read the cookie state once it's
      // changed, so even though we've updated the cookie, the App will still
      // think that the user isn't logged in and will show the login page.
      window.location.reload();
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
