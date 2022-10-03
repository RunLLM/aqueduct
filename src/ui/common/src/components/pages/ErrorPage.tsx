import { Box } from '@mui/material';
import Link from '@mui/material/Link';
import Typography from '@mui/material/Typography';
import React from 'react';

import UserProfile from '../../utils/auth';
import DefaultLayout from '../layouts/default';
import { LayoutProps } from './types';

type ErrorPageProps = {
  user?: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const ErrorPage: React.FC<ErrorPageProps> = ({
  user,
  Layout = DefaultLayout,
}) => {
  const contents = (
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
          <Typography variant="body1" sx={{ fontSize: '20px' }}>
            Something went wrong.
          </Typography>
          <Typography
            variant="body1"
            sx={{ textAlign: 'center', fontSize: '15px', mt: '16px' }}
          >
            If this problem continues to persist, you make a post in our{' '}
            <Link href="https://join.slack.com/t/aqueductusers/shared_invite/zt-11hby91cx-cpmgfK0qfXqEYXv25hqD6A">
              Slack community
            </Link>{' '}
            or contact our team directly at{' '}
            <Link href="mailto:support@aqueducthq.com">
              support@aqueducthq.com
            </Link>
            .
          </Typography>
        </Box>
      </Box>
    </Box>
  );

  if (user) {
    return <Layout user={user}>{contents}</Layout>;
  }
  return contents;
};

export default ErrorPage;
