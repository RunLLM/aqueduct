import { Alert, CircularProgress, Snackbar, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { IntegrationCard } from '../../components/integrations/cards/card';
import { useIntegrationsWorkflowsGetQuery } from '../../handlers/AqueductApi';
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
  const [showWorkflowsFetchErrorToast, setShowWorkflowsFetchErrorToast] =
    useState(false);
  const handleWorkflowsFetchErrorToastClose = () => {
    setShowWorkflowsFetchErrorToast(false);
  };

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

  // For each integration, count the number of workflows that use it.
  // Fetch the number of workflows for each integration.
  const {
    data: workflowsByIntegration,
    error: fetchWorkflowsError,
    isLoading,
  } = useIntegrationsWorkflowsGetQuery({ apiKey: user.apiKey });
  if (isLoading) {
    return <CircularProgress />;
  }
  if (fetchWorkflowsError) {
    setShowWorkflowsFetchErrorToast(true);
  }

  // Had to move this down here because react hooks don't like it when there are early returns
  // in front of them.
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
      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showWorkflowsFetchErrorToast}
        onClose={handleWorkflowsFetchErrorToastClose}
        key={'integrations-dialog-success-snackbar'}
        autoHideDuration={6000}
      >
        <Alert
          onClose={handleWorkflowsFetchErrorToastClose}
          severity="error"
          sx={{ width: '100%' }}
        >
          Unexpected error occurred when fetching the workflows associated with
          the integrations.
        </Alert>
      </Snackbar>

      <Typography variant="h6">{connectedIntegrationType}</Typography>
      <Box
        sx={{
          display: 'flex',
          flexWrap: 'wrap',
          alignItems: 'flex-start',
        }}
      >
        {[...integrations]
          .sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1))
          .map((integration, idx) => {
            // Leave this empty if there was an error while fetching workflows.
            let numWorkflowsUsingMsg = '';
            const numWorkflowsUsing =
              workflowsByIntegration[integration.id].length;
            if (!fetchWorkflowsError) {
              if (numWorkflowsUsing > 0) {
                numWorkflowsUsingMsg = `Used by ${numWorkflowsUsing} ${
                  numWorkflowsUsing === 1 ? 'workflow' : 'workflows'
                }`;
              } else {
                numWorkflowsUsingMsg = 'Not currently in use';
              }
            }

            return (
              <Box key={idx} sx={{ mx: 1, my: 1 }}>
                <Link
                  underline="none"
                  color="inherit"
                  href={`${getPathPrefix()}/integration/${integration.id}`}
                >
                  <Card>
                    <IntegrationCard
                      integration={integration}
                      numWorkflowsUsingMsg={numWorkflowsUsingMsg}
                    />
                  </Card>
                </Link>
              </Box>
            );
          })}
      </Box>
    </Box>
  );
};
