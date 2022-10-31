import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import Link from '@mui/material/Link';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { IntegrationCard } from '../../components/integrations/cards/card';
import { handleLoadIntegrations } from '../../reducers/integrations';
import { AppDispatch, RootState } from '../../stores/store';
import { UserProfile } from '../../utils/auth';
import { getPathPrefix } from '../../utils/getPathPrefix';
import { Card } from '../layouts/card';

type ConnectedIntegrationsProps = {
  user: UserProfile;
  forceLoad: boolean;
};

export const ConnectedIntegrations: React.FC<ConnectedIntegrationsProps> = ({
  user,
  forceLoad,
}) => {
  const dispatch: AppDispatch = useDispatch();

  useEffect(() => {
    dispatch(
      handleLoadIntegrations({ apiKey: user.apiKey, forceLoad: forceLoad })
    );
  }, [dispatch, forceLoad, user.apiKey]);

  const integrations = useSelector((state: RootState) =>
    Object.values(state.integrationsReducer.integrations)
  );
  if (!integrations) {
    return null;
  }

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'flex-start',
      }}
    >
      {[...integrations]
        .sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1))
        .map((integration, idx) => {
          return (
            <Box key={idx} sx={{ width: '90%', maxWidth: '1000px' }}>
              <Link
                underline="none"
                color="inherit"
                href={`${getPathPrefix()}/integration/${integration.id}`}
              >
                <Card sx={{ my: 2 }}>
                  <IntegrationCard integration={integration} />
                </Card>
              </Link>

              {idx < integrations.length - 1 && <Divider />}
            </Box>
          );
        })}
    </Box>
  );
};
