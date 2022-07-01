import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';
import { CopyBlock, github } from 'react-code-blocks';

import UserProfile from '../../utils/auth';
import { useAqueductConsts } from '../hooks/useAqueductConsts';
import DefaultLayout from '../layouts/default';

type AccountPageProps = {
  user: UserProfile;
};

const AccountPage: React.FC<AccountPageProps> = ({ user }) => {
  // Set the title of the page on page load.
  useEffect(() => {
    document.title = 'Account | Aqueduct';
  }, []);

  const { apiAddress } = useAqueductConsts();
  const serverAddress = apiAddress ? `${apiAddress}` : '<server address>';
  const apiConnectionSnippet = `import aqueduct
client = aqueduct.Client(
    "${user.apiKey}",
    "${serverAddress}"
)`;
  const maxContentWidth = '600px';

  return (
    <DefaultLayout user={user}>
      <Typography variant="h2" gutterBottom component="div">
        Account Overview
      </Typography>

      <Typography variant="h5" sx={{ mt: 3 }}>
        API Key
      </Typography>
      <Box sx={{ my: 1 }}>
        <code>{user.apiKey}</code>
      </Box>

      <Box
        sx={{
          display: 'flex',
          flexDirection: 'column',
          width: maxContentWidth,
        }}
      >
        <Typography variant="body1" sx={{ fontWeight: 'bold', mr: '8px' }}>
          Python SDK Connection Snippet
        </Typography>
        <Box
          sx={{
            span: { padding: '0 !important' },
          }}
        >
          <CopyBlock
            text={apiConnectionSnippet}
            language="python"
            showLineNumbers={false}
            theme={github}
          />
        </Box>
      </Box>
    </DefaultLayout>
  );
};

export default AccountPage;
