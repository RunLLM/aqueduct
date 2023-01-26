//import styled from '@emotion/styled';
import { styled } from '@mui/material/styles';

import { faSlack } from '@fortawesome/free-brands-svg-icons';
import {
  faBook,
  faEnvelope,
  faThumbsUp,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';

import { theme } from '../../styles/theme/theme';
import UserProfile from '../../utils/auth';
import { getPathPrefix } from '../../utils/getPathPrefix';

const InfoCard = styled(Box)({
  ['&.MuiBox-root']: {
    display: 'flex',
    borderWidth: '2px',
    borderColor: '#002F5E',
    borderRadius: '5px',
    borderStyle: 'solid',
    width: '100%',
    marginTop: '8px',
    alignItems: 'center',
    padding: '16px 8px 16px 8px',
    '&:hover': {
      backgroundColor: theme.palette.blue[50],
    },
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

  const waveEmoji = () => {
    const emojiElement = document.getElementById('greet-emoji');
    emojiElement.style.transitionDuration = '.4s';
    emojiElement.style.transformOrigin = 'right bottom';

    const delayIncrement = 400;
    let delay = 100;
    for (let i = 0; i < 8; i++) {
      if (i % 2 === 0) {
        setTimeout(
          () => (emojiElement.style.transform = 'rotate(30deg)'),
          delay
        );
      } else {
        setTimeout(
          () => (emojiElement.style.transform = 'rotate(-30deg)'),
          delay
        );
      }

      delay += delayIncrement;
    }

    setTimeout(() => (emojiElement.style.transform = 'rotate(0deg)'), delay);
  };

  useEffect(waveEmoji, []);

  return (
    <Box
      sx={{
        width: '100%',
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
          maxWidth: '750px',
          mt: '32px',
        }}
      >
        {user && (
          <InfoCard sx={{ mb: '32px' }} onMouseEnter={() => waveEmoji()}>
            <Box
              sx={{
                display: 'flex',
                flexDirection: 'column',
                textAlign: 'center',
                alignItems: 'center',
              }}
            >
              <Typography
                sx={{ width: 'min-content' }}
                variant="h4"
                id="greet-emoji"
              >
                👋
              </Typography>
              <Typography variant="h4">{greeting}</Typography>

              <Typography variant="h6">
                Here&apos;s how you can get started.
              </Typography>

              <Box sx={{ textAlign: 'left', mr: '16px' }}>
                <ol>
                  <li>
                    <Typography variant="body1">
                      First go to the{' '}
                      <Link href={`${getPathPrefix()}/integrations`}>
                        integrations
                      </Link>{' '}
                      page and connect a database. If you don&apos;t have a
                      database handy, you can use the <code>aqueduct_demo</code>{' '}
                      database -- see the documentation{' '}
                      <Link href="https://docs.aqueducthq.com/example-workflows/demo-data-warehouse">
                        here
                      </Link>
                      .
                    </Typography>
                  </li>

                  <li>
                    <Typography variant="body1">
                      Create your first workflow -- see our{' '}
                      <Link href="https://docs.aqueducthq.com/quickstart-guide">
                        {' '}
                        Quickstart Guide{' '}
                      </Link>{' '}
                      for an example.
                    </Typography>
                  </li>

                  <li>
                    <Typography variant="body1">
                      Go to the{' '}
                      <Link href={`${getPathPrefix()}/workflows`}>
                        workflows
                      </Link>{' '}
                      page to see a visualization of the workflow you just
                      created.
                    </Typography>
                  </li>
                </ol>
              </Box>
            </Box>
          </InfoCard>
        )}

        <InfoCard>
          <Box sx={{ width: '80px', px: 2, fontSize: '48px' }}>
            <FontAwesomeIcon icon={faBook} color="#002F5E" />
          </Box>
          <Typography variant="body1" sx={{ fontSize: '20px' }}>
            {`If you still have questions, you can find our documentation `}
            <Link href="https://docs.aqueducthq.com">here</Link>.
          </Typography>
        </InfoCard>

        <InfoCard>
          <Box sx={{ width: '80px', px: 2, fontSize: '48px' }}>
            <FontAwesomeIcon icon={faEnvelope} color="#002F5E" />
          </Box>
          <Typography variant="body1" sx={{ fontSize: '20px' }}>
            {`Have any other questions? You can reach our team via email at `}
            <Link href="mailto:support@aqueducthq.com">
              support@aqueducthq.com
            </Link>
            .
          </Typography>
        </InfoCard>

        <InfoCard>
          <Box sx={{ width: '80px', px: 2, fontSize: '48px' }}>
            <FontAwesomeIcon icon={faSlack} color="#002F5E" />
          </Box>
          <Typography variant="body1" sx={{ fontSize: '20px' }}>
            {`You can also join our Slack community `}
            <Link href="https://join.slack.com/t/aqueductusers/shared_invite/zt-11hby91cx-cpmgfK0qfXqEYXv25hqD6A">
              here
            </Link>
            .
          </Typography>
        </InfoCard>

        <InfoCard>
          <Box sx={{ width: '80px', px: 2, fontSize: '48px' }}>
            <FontAwesomeIcon icon={faThumbsUp} color="#002F5E" />
          </Box>
          <Typography variant="body1" sx={{ fontSize: '20px' }}>
            {`Have feedback? Let us know how we're doing `}
            <Link href="https://forms.gle/Ef5hvT35d7j27YqV6">here</Link>.
          </Typography>
        </InfoCard>
      </Box>
    </Box>
  );
};

export default GettingStartedTutorial;
