import { Alert, Snackbar, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useLocation } from 'react-router-dom';

import { useStorageMigrationListQuery } from '../../../handlers/AqueductApi';
import { handleGetServerConfig } from '../../../handlers/getServerConfig';
import { StorageMigrationResponse } from '../../../handlers/responses/storageMigration';
import { RootState } from '../../../stores/store';
import { theme } from '../../../styles/theme/theme';
import UserProfile from '../../../utils/auth';
import { ResourceCategories } from '../../../utils/resources';
import { LoadingStatus, LoadingStatusEnum } from '../../../utils/shared';
import SupportedResources from '../../../utils/SupportedResources';
import DefaultLayout from '../../layouts/default';
import { BreadcrumbLink } from '../../layouts/NavBar';
import AddResources from '../../resources/addResources';
import { ConnectedResources } from '../../resources/connectedResources';
import { ConnectedResourceType } from '../../resources/connectedResourceType';
import MetadataStorageInfo from '../account/MetadataStorageInfo';
import { LayoutProps } from '../types';

type Props = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

type resourcesNavigateState = {
  deleteResourceStatus: LoadingStatus;
  deleteResourceName: string;
};

const ResourcesPage: React.FC<Props> = ({ user, Layout = DefaultLayout }) => {
  const location = useLocation();

  const serverConfig = useSelector(
    (state: RootState) => state.serverConfigReducer
  );
  const dispatch = useDispatch();
  useEffect(() => {
    async function fetchServerConfig() {
      if (user) {
        await dispatch(handleGetServerConfig({ apiKey: user.apiKey }));
      }
    }

    fetchServerConfig();
  }, [user]);

  useEffect(() => {
    document.title = 'Resources | Aqueduct';
  }, []);

  let deleteResourceName = '';
  let forceLoad = false;

  const [showDeleteResourceSuccessToast, setShowDeleteResourceSuccessToast] =
    useState(false);

  if (location.state && location.state !== undefined) {
    const navState = location.state as resourcesNavigateState;
    deleteResourceName = navState.deleteResourceName;
    if (!showDeleteResourceSuccessToast) {
      setShowDeleteResourceSuccessToast(
        navState.deleteResourceStatus.loading === LoadingStatusEnum.Succeeded
      );
    }

    // Reload resources because deleted
    forceLoad = true;
  }

  // If the last storage migration failed, display the error message.
  const { data, error, isLoading } = useStorageMigrationListQuery({
    apiKey: user.apiKey,
    limit: '1', // only fetch the latest result.
  });
  const lastMigration = data as StorageMigrationResponse[];
  let lastFailedFormattedTimestamp: string | undefined = undefined;
  if (lastMigration && lastMigration[0].execution_state.status === 'failed') {
    const date = new Date(
      lastMigration[0].execution_state.timestamps.registered_at
    );
    lastFailedFormattedTimestamp = date.toLocaleString();
  }

  return (
    <Layout
      breadcrumbs={[BreadcrumbLink.HOME, BreadcrumbLink.INTEGRATIONS]}
      user={user}
    >
      <Box paddingBottom={4}>
        <Typography variant="h5" marginBottom={2}>
          Available Resources
        </Typography>
        <ConnectedResources
          user={user}
          forceLoad={forceLoad}
          connectedResourceType={ConnectedResourceType.Compute}
        />
        <ConnectedResources
          user={user}
          forceLoad={forceLoad}
          connectedResourceType={ConnectedResourceType.Data}
        />
        <ConnectedResources
          user={user}
          forceLoad={forceLoad}
          connectedResourceType={ConnectedResourceType.Other}
        />

        <Box>
          <Typography variant="h6" marginY={2}>
            Artifact Storage
          </Typography>
          <MetadataStorageInfo serverConfig={serverConfig.config} />
          {!isLoading && lastFailedFormattedTimestamp && (
            <Box>
              <Typography
                variant="body2"
                fontWeight="fontWeightRegular"
                marginTop={2}
                marginBottom={1}
              >
                The last artifact storage migration, which started at{' '}
                {lastFailedFormattedTimestamp}, has failed! As a result, the
                artifact storage has not changed from `
                {serverConfig.config?.storageConfig.resource_name}`.
              </Typography>
              <Box
                sx={{
                  borderRadius: 2,
                  backgroundColor: theme.palette.red[100],
                  color: theme.palette.red[600],
                  p: 2,
                  paddingBottom: '16px',
                  paddingTop: '16px',
                  height: 'fit-content',
                }}
              >
                <pre style={{ margin: '0px' }}>
                  {`${lastMigration[0].execution_state.error.tip}\n\n${lastMigration[0].execution_state.error.context}`}
                </pre>
              </Box>
            </Box>
          )}
        </Box>

        <Box marginY={2}>
          <Divider />
        </Box>

        <Typography variant="h5" marginY={2}>
          Add New Resources
        </Typography>

        <Typography variant="h6" marginY={2}>
          Compute
        </Typography>
        <AddResources
          user={user}
          category={ResourceCategories.COMPUTE}
          supportedResources={SupportedResources}
        />
        <Typography variant="h6" marginY={2}>
          Data
        </Typography>
        <AddResources
          user={user}
          category={ResourceCategories.DATA}
          supportedResources={SupportedResources}
        />
        <Typography variant="h6" marginY={2}>
          Container Registry
        </Typography>
        <AddResources
          user={user}
          category={ResourceCategories.CONTAINER_REGISTRY}
          supportedResources={SupportedResources}
        />

        <Typography variant="h6" marginY={2}>
          Notifications
        </Typography>
        <AddResources
          user={user}
          category={ResourceCategories.NOTIFICATION}
          supportedResources={SupportedResources}
        />
      </Box>

      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showDeleteResourceSuccessToast}
        key={'workflowheader-delete-success-error-snackbar'}
        autoHideDuration={6000}
        onClose={() => {
          setShowDeleteResourceSuccessToast(false);
          location.state = undefined;
        }}
      >
        <Alert severity="success" sx={{ width: '100%' }}>
          {`Successfully deleted ${deleteResourceName}`}
        </Alert>
      </Snackbar>
    </Layout>
  );
};

export default ResourcesPage;
