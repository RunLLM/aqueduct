import { faSlack } from '@fortawesome/free-brands-svg-icons';
import {
  faBook,
  faEnvelope,
  faThumbsUp,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import { styled } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import React from 'react';

import UserProfile from '../../utils/auth';
import { getPathPrefix } from '../../utils/getPathPrefix';

const sharedCardStyling = {
  display: 'flex',
  backgroundColor: 'gray.50',
  borderRadius: 3,
  paddingX: 1,
  paddingY: 2,
};

const InfoCard = styled(Box)({
  ['&.MuiBox-root']: {
    ...sharedCardStyling,
    marginTop: '8px',
    height: '125px',
    width: '50%',
    display: 'flex',
    alignItems: 'center',
    minWidth: '450px',
    maxWidth: '750px',
  },
});

type GettingStartedTutorialProps = {
  user: UserProfile;
};

const GettingStartedTutorial: React.FC<GettingStartedTutorialProps> = ({
  user,
}) => {
  let greeting = `Welcome ${user.given_name ?? user.email}!`;
  if (user.email === 'default' || !user.given_name) {
    greeting = 'Welcome to Aqueduct!';
  }

  return (
    <Box
      sx={{
        width: '100%',
        height: '100%',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        flexDirection: 'column',
      }}
    >
      <Box
        sx={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          width: '50%',
          backgroundColor: 'white',
          minWidth: '500px',
        }}
      >
        {user && (
          <Box
            sx={{
              ...sharedCardStyling,
              flexDirection: 'column',
              textAlign: 'center',
              maxWidth: '750px',
              marginY: '32px',
            }}
          >
            <Typography variant="h4">👋 {greeting}</Typography>

            <Typography variant="h6">
              Here&apos;s how you can get started with Aqueduct.
            </Typography>

            <Box sx={{ textAlign: 'left', mr: '16px' }}>
              <ol>
                <li>
                  <Typography variant="body1">
                    First go to the{' '}
                    <Link href={`${getPathPrefix()}/integrations`}>
                      integrations
                    </Link>{' '}
                    page and connect a database. (If you don&apos;t have a
                    database handy, you can use the <code>aqueduct_demo</code>{' '}
                    database -- see the documentation{' '}
                    <Link href="https://docs.aqueducthq.com/example-workflows/demo-data-warehouse">
                      here
                    </Link>
                    .)
                  </Typography>
                </li>

                <li>
                  <Typography variant="body1">
                    Install our{' '}
                    <Link href="https://github.com/aqueducthq/aqueduct/blob/main/sdk">
                      Python SDK
                    </Link>
                    .
                  </Typography>
                </li>

                <li>
                  <Typography variant="body1">
                    Create your first workflow from the Python SDK -- see our{' '}
                    <Link href="https://github.com/aqueducthq/aqueduct/blob/main/sdk/examples/churn_prediction/Build%20and%20Deploy%20Churn%20Ensemble.ipynb">
                      {' '}
                      churn example{' '}
                    </Link>{' '}
                    for some inspiration.
                  </Typography>
                </li>

                <li>
                  <Typography variant="body1">
                    Go to the{' '}
                    <Link href={`${getPathPrefix()}/workflows`}>workflows</Link>{' '}
                    page to see a visualization of the workflow you just
                    created.
                  </Typography>
                </li>
              </ol>
            </Box>
          </Box>
        )}
      </Box>

      {/* `styled` currently doesn't respect `backgroundColor` for some reason, so we have to set it explicitly here.. */}
      <InfoCard sx={{ backgroundColor: 'gray.50' }}>
        <Box sx={{ width: '80px', px: 2, fontSize: '48px' }}>
          <FontAwesomeIcon icon={faBook} />
        </Box>
        <Typography variant="h6">
          {`If you still have questions, you can find our documentation `}
          <Link href="https://docs.aqueducthq.com">here</Link>.
        </Typography>
      </InfoCard>

      <InfoCard sx={{ backgroundColor: 'gray.50' }}>
        <Box sx={{ width: '80px', px: 2, fontSize: '48px' }}>
          <FontAwesomeIcon icon={faEnvelope} />
        </Box>
        <Typography variant="h6">
          {`Have any other questions? You can reach our team via email at `}
          <Link href="mailto:support@aqueducthq.com">
            support@aqueducthq.com
          </Link>
          .
        </Typography>
      </InfoCard>

      <InfoCard sx={{ backgroundColor: 'gray.50' }}>
        <Box sx={{ width: '80px', px: 2, fontSize: '48px' }}>
          <FontAwesomeIcon icon={faSlack} />
        </Box>
        <Typography variant="h6">
          {`You can also join our Slack community `}
          <Link href="https://join.slack.com/t/aqueductusers/shared_invite/zt-11hby91cx-cpmgfK0qfXqEYXv25hqD6A">
            here
          </Link>
          .
        </Typography>
      </InfoCard>

      <InfoCard sx={{ backgroundColor: 'gray.50' }}>
        <Box sx={{ width: '80px', px: 2, fontSize: '48px' }}>
          <FontAwesomeIcon icon={faThumbsUp} />
        </Box>
        <Typography variant="h6">
          {`Have feedback? Let us know how we're doing `}
          <Link href="https://forms.gle/Ef5hvT35d7j27YqV6">here</Link>.
        </Typography>
      </InfoCard>
    </Box>
  );
};

export default GettingStartedTutorial;
