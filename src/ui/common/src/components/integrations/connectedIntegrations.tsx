import { Typography } from '@mui/material';
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
import {
  Integration,
  IntegrationCategories,
  SupportedIntegrations,
} from '../../utils/integrations';
import { Card } from '../layouts/card';
import { ConnectedIntegrationType } from './connectedIntegrationType';

type ConnectedIntegrationsProps = {
  user: UserProfile;
  forceLoad: boolean;

  // This filters the displayed integrations to only those of the given type.
  connectedIntegrationType: ConnectedIntegrationType;
};

export const ConnectedIntegrations: React.FC<ConnectedIntegrationsProps> = ({
  user,
  forceLoad,
  connectedIntegrationType,
}) => {
  const dispatch: AppDispatch = useDispatch();

  useEffect(() => {
    dispatch(
      handleLoadIntegrations({ apiKey: user.apiKey, forceLoad: forceLoad })
    );
  }, [dispatch, forceLoad, user.apiKey]);

  const integrationToConnectedIntegrationType = (integration: Integration) => {
    if (
      SupportedIntegrations[integration.service].category ===
      IntegrationCategories.DATA
    ) {
      return ConnectedIntegrationType.Data;
    } else if (
      SupportedIntegrations[integration.service].category ===
        IntegrationCategories.COMPUTE ||
      SupportedIntegrations[integration.service].category ===
        IntegrationCategories.CLOUD
    ) {
      return ConnectedIntegrationType.Compute;
    } else {
      return ConnectedIntegrationType.Other;
    }
  };

  const integrations = useSelector((state: RootState) =>
    Object.values(state.integrationsReducer.integrations).filter(
      (integration) =>
        integrationToConnectedIntegrationType(integration) ===
        connectedIntegrationType
    )
  );
  if (!integrations) {
    return null;
  }

  // Do not show the "Other" section if there are no "Other" integrations.
  if (
    integrations.length === 0 &&
    connectedIntegrationType === ConnectedIntegrationType.Other
  ) {
    return null;
  }

  return (
    <Box>
      <Typography variant="h6">{connectedIntegrationType}</Typography>

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
    </Box>
  );
};
