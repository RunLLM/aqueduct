import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';
import React, { ChangeEvent, useCallback, useEffect, useState } from 'react';
import { useCookies } from 'react-cookie';
import { useSearchParams, useNavigate } from 'react-router-dom';

import fetchUser from '../../utils/fetchUser';
import { getPathPrefix } from '../../utils/getPathPrefix';
import { Button } from '../primitives/Button.styles';

const apiKeyQueryParam = 'apiKey';

export const LoginPage: React.FC = () => {
  const navigate = useNavigate()
  useEffect(() => {
    //navigate('https://5729345786.propelauthtest.com', { replace: true })
    window.location.href = 'https://5729345786.propelauthtest.com'
  }, [])

  return null;
  /*
  useEffect(() => {
    document.title = 'Login | Aqueduct';
  }, []);

  const [searchParams] = useSearchParams();
  const [cookies, setCookie] = useCookies(['aqueduct-api-key']);
  const [validationError, setValidationError] = useState<boolean>(false);
  const [errorMsg, setErrorMsg] = useState<string>('');

  // The cookies library is kinda dumb and sometimes returns the word
  // undefined as a string rather than returning an undefined value, hence the
  // extra check here.
  const [apiKey, setApiKey] = useState<string>(
    cookies['aqueduct-api-key'] && cookies['aqueduct-api-key'] !== 'undefined'
      ? cookies['aqueduct-api-key']
      : ''
  );

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

  const onGetStartedClicked = useCallback(
    async (key: string) => {
      const { success } = await fetchUser(key);

      if (!success) {
        setValidationError(true);
        setErrorMsg(
          'Invalid API Key. You can find your API Key by running `aqueduct apikey` on the machine where Aqueduct is running.'
        );
      } else {
        setCookie('aqueduct-api-key', key, { path: '/' });
        await new Promise((r) => setTimeout(r, 100));
        setValidationError(false);

        // Once we validate, we force a reload of the page. This is because React
        // doesn't give us an easy way to read the cookie state once it's
        // changed, so even though we've updated the cookie, the App will still
        // think that the user isn't logged in and will show the login page.
        window.location.assign(getPathPrefix());
      }
    },
    [setCookie]
  );

  // On page load, check if there's a query parameter with the API key. If there
  // is, then we automatically try to login with that API key.
  useEffect(() => {
    const key = searchParams.get(apiKeyQueryParam);

    if (key && key.length > 0) {
      setApiKey(key);
      onGetStartedClicked(key);
    }
  }, [onGetStartedClicked, searchParams]);

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
            value={apiKey ?? ''}
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
          onClick={() => onGetStartedClicked(apiKey)}
          sx={{ marginTop: 1 }}
          fullWidth={true}
          color="primary"
          variant="contained"
        >
          Get Started
        </Button>
      </Box>
    </Box>
  );*/
};

export default LoginPage;
