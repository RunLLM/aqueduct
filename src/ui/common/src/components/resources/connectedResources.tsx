import { CircularProgress, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { ResourceCard } from '../../components/resources/cards/card';
import { useResourcesWorkflowsGetQuery } from '../../handlers/AqueductApi';
import { handleLoadResources } from '../../reducers/resources';
import { AppDispatch, RootState } from '../../stores/store';
import { UserProfile } from '../../utils/auth';
import { getPathPrefix } from '../../utils/getPathPrefix';
import { Resource, ResourceCategories } from '../../utils/resources';
import SupportedResources from '../../utils/SupportedResources';
import { Card } from '../layouts/card';
import { ConnectedResourceType } from './connectedResourceType';
import { ErrorSnackbar } from './errorSnackbar';
import { getNumWorkflowsUsingMessage } from './numWorkflowsUsingMsg';

type ConnectedResourcesProps = {
  user: UserProfile;
  forceLoad: boolean;

  // This filters the displayed resources to only those of the given type.
  connectedResourceType: ConnectedResourceType;
};

export const ConnectedResources: React.FC<ConnectedResourcesProps> = ({
  user,
  forceLoad,
  connectedResourceType,
}) => {
  const dispatch: AppDispatch = useDispatch();

  useEffect(() => {
    dispatch(
      handleLoadResources({ apiKey: user.apiKey, forceLoad: forceLoad })
    );
  }, [dispatch, forceLoad, user.apiKey]);

  const resourceToConnectedResourceType = (resource: Resource) => {
    if (
      SupportedResources[resource.service].category === ResourceCategories.DATA
    ) {
      return ConnectedResourceType.Data;
    } else if (
      SupportedResources[resource.service].category ===
        ResourceCategories.COMPUTE ||
      SupportedResources[resource.service].category === ResourceCategories.CLOUD
    ) {
      return ConnectedResourceType.Compute;

      // The "Artifact Storage" is currently only used to filter out the 'Filesystem' resource
      // from the connected resources.
    } else if (
      SupportedResources[resource.service].category ===
      ResourceCategories.ARTIFACT_STORAGE
    ) {
      return ConnectedResourceType.ArtifactStorage;
    } else {
      return ConnectedResourceType.Other;
    }
  };

  const resources = useSelector((state: RootState) =>
    Object.values(state.resourcesReducer.resources).filter(
      (resource: Resource) =>
        resourceToConnectedResourceType(resource) === connectedResourceType
    )
  );

  // For each resource, count the number of workflows that use it.
  // Fetch the number of workflows for each resource.
  const {
    data: workflowAndDagIDsByResource,
    error: fetchWorkflowsError,
    isLoading,
  } = useResourcesWorkflowsGetQuery({ apiKey: user.apiKey });

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
    connectedResourceType === ConnectedResourceType.Other
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

      <Typography variant="h6">{connectedResourceType}</Typography>
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
              resource.id in workflowAndDagIDsByResource
            ) {
              numWorkflowsUsingMsg = getNumWorkflowsUsingMessage(
                workflowAndDagIDsByResource[resource.id].length
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
                    <ResourceCard
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
