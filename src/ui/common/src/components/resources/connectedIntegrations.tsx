import { CircularProgress, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { IntegrationCard } from '../../components/resources/cards/card';
import { useIntegrationsWorkflowsGetQuery } from '../../handlers/AqueductApi';
import { handleLoadIntegrations } from '../../reducers/resources';
import { AppDispatch, RootState } from '../../stores/store';
import { UserProfile } from '../../utils/auth';
import { getPathPrefix } from '../../utils/getPathPrefix';
import { Integration, IntegrationCategories } from '../../utils/resources';
import SupportedIntegrations from '../../utils/SupportedIntegrations';
import { Card } from '../layouts/card';
import { ConnectedIntegrationType } from './connectedIntegrationType';
import { ErrorSnackbar } from './errorSnackbar';
import { getNumWorkflowsUsingMessage } from './numWorkflowsUsingMsg';

type ConnectedIntegrationsProps = {
  user: UserProfile;
  forceLoad: boolean;

  // This filters the displayed resources to only those of the given type.
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

  const resourceToConnectedIntegrationType = (resource: Integration) => {
    if (
      SupportedIntegrations[resource.service].category ===
      IntegrationCategories.DATA
    ) {
      return ConnectedIntegrationType.Data;
    } else if (
      SupportedIntegrations[resource.service].category ===
        IntegrationCategories.COMPUTE ||
      SupportedIntegrations[resource.service].category ===
        IntegrationCategories.CLOUD
    ) {
      return ConnectedIntegrationType.Compute;

      // The "Artifact Storage" is currently only used to filter out the 'Filesystem' resource
      // from the connected resources.
    } else if (
      SupportedIntegrations[resource.service].category ===
      IntegrationCategories.ARTIFACT_STORAGE
    ) {
      return ConnectedIntegrationType.ArtifactStorage;
    } else {
      return ConnectedIntegrationType.Other;
    }
  };

  const resources = useSelector((state: RootState) =>
    Object.values(state.resourcesReducer.resources).filter(
      (resource: Integration) =>
        resourceToConnectedIntegrationType(resource) ===
        connectedIntegrationType
    )
  );

  // For each resource, count the number of workflows that use it.
  // Fetch the number of workflows for each resource.
  const {
    data: workflowAndDagIDsByIntegration,
    error: fetchWorkflowsError,
    isLoading,
  } = useIntegrationsWorkflowsGetQuery({ apiKey: user.apiKey });

  if (isLoading) {
    return <CircularProgress />;
  }

  // Had to move this down here because react hooks don't like it when there are early returns
  // in front of them.
  if (!resources) {
    return null;
  }

  // Do not show the "Other" section if there are no "Other" resources.
  if (
    resources.length === 0 &&
    connectedIntegrationType === ConnectedIntegrationType.Other
  ) {
    return null;
  }

  return (
    <Box>
      <ErrorSnackbar
        shouldShow={fetchWorkflowsError !== undefined}
        errMsg={
          'Unexpected error occurred when fetching the workflows associated with the resources. Please try again.'
        }
      />

      <Typography variant="h6">{connectedIntegrationType}</Typography>
      <Box
        sx={{
          display: 'flex',
          flexWrap: 'wrap',
          alignItems: 'flex-start',
        }}
      >
        {[...resources]
          // This is a temporary fix to hide the auto-generated on-demand k8s resource card.
          // This also filters out any Conda resource, since that is merged in with the Aqueduct Server card.
          .filter(
            (resource) =>
              !resource.name.endsWith(':aqueduct_ondemand_k8s') &&
              resource.service != 'Conda'
          )
          .sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1))
          .map((resource, idx) => {
            let numWorkflowsUsingMsg = '';
            if (
              !fetchWorkflowsError &&
              resource.id in workflowAndDagIDsByIntegration
            ) {
              numWorkflowsUsingMsg = getNumWorkflowsUsingMessage(
                workflowAndDagIDsByIntegration[resource.id].length
              );
            }

            return (
              <Box key={idx} sx={{ mx: 1, my: 1 }}>
                <Link
                  underline="none"
                  color="inherit"
                  href={`${getPathPrefix()}/resource/${resource.id}`}
                >
                  <Card>
                    <IntegrationCard
                      resource={resource}
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
