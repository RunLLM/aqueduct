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
  user, forceLoad,
}) => {
  const dispatch: AppDispatch = useDispatch();

  useEffect(() => {
    dispatch(handleLoadIntegrations({ apiKey: user.apiKey, forceLoad: forceLoad }));
  }, []);

  const integrations = useSelector(
    (state: RootState) => state.integrationsReducer.integrations
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
        my: 1,
      }}
    >
      {[...integrations]
        .sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1))
        .map((integration, idx) => {
          return (
            <Box key={idx}>
              <Link
                underline="none"
                color="inherit"
                href={`${getPathPrefix()}/integration/${integration.id}`}
              >
                <Card sx={{ marginY: 2 }}>
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
