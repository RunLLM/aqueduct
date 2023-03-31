import { Alert, Snackbar, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import React, { useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';

import { BreadcrumbLink } from '../../../components/layouts/NavBar';
import UserProfile from '../../../utils/auth';
import {
  IntegrationCategories,
  SupportedIntegrations,
} from '../../../utils/integrations';
import { LoadingStatus, LoadingStatusEnum } from '../../../utils/shared';
import AddIntegrations from '../../integrations/addIntegrations';
import { ConnectedIntegrations } from '../../integrations/connectedIntegrations';
import DefaultLayout from '../../layouts/default';
import { LayoutProps } from '../types';
import MetadataStorageInfo from "../account/MetadataStorageInfo";
import {useDispatch, useSelector} from "react-redux";
import {RootState} from "../../../stores/store";
import {handleGetServerConfig} from "../../../handlers/getServerConfig";
import {useStorageMigrationListQuery} from "../../../handlers/AqueductApi";
import {theme} from "../../../styles/theme/theme";
import {StorageMigrationResponse} from "../../../handlers/responses/storageMigration";

type Props = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

type integrationsNavigateState = {
  deleteIntegrationStatus: LoadingStatus;
  deleteIntegrationName: string;
};

const IntegrationsPage: React.FC<Props> = ({
  user,
  Layout = DefaultLayout,
}) => {
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
    document.title = 'Integrations | Aqueduct';
  }, []);

  let deleteIntegrationName = '';
  let forceLoad = false;

  const [
    showDeleteIntegrationSuccessToast,
    setShowDeleteIntegrationSuccessToast,
  ] = useState(false);

  if (location.state && location.state !== undefined) {
    const navState = location.state as integrationsNavigateState;
    deleteIntegrationName = navState.deleteIntegrationName;
    if (!showDeleteIntegrationSuccessToast) {
      setShowDeleteIntegrationSuccessToast(
        navState.deleteIntegrationStatus.loading === LoadingStatusEnum.Succeeded
      );
    }

    // Reload integrations because deleted
    forceLoad = true;
  }

  // If the last storage migration failed, display the error message.
  const { data, error, isLoading } = useStorageMigrationListQuery(
      {
        apiKey: user.apiKey,
        limit: '1', // only fetch the latest result.
      }
  )
  const lastMigration = data as StorageMigrationResponse[]
  let lastFailedFormattedTimestamp: string | undefined = undefined;
  if (lastMigration && lastMigration[0].execution_state.status === "failed") {
    const date = new Date(lastMigration[0].execution_state.timestamps.registered_at);
    const year = date.getFullYear();
    const month = date.getMonth() + 1; // Note: getMonth() returns 0-based index
    const day = date.getDate();
    const hour = date.getHours() % 12; // Convert to 12-hour format
    const minute = date.getMinutes();
    const amPm = date.getHours() >= 12 ? "pm" : "am"; // Determine AM/PM based on hour
    lastFailedFormattedTimestamp = `${year}-${month < 10 ? "0" : ""}${month}-${day < 10 ? "0" : ""}${day} ${hour}:${minute < 10 ? "0" : ""}${minute}${amPm}`;
  }

  return (
    <Layout
      breadcrumbs={[BreadcrumbLink.HOME, BreadcrumbLink.INTEGRATIONS]}
      user={user}
    >
      <Box>
        <Box>
          <Typography variant="h5" marginBottom={2}>
            Add an Integration
          </Typography>
          <Typography variant="h6" marginY={2}>
            Cloud
          </Typography>
          <AddIntegrations
            user={user}
            category={IntegrationCategories.CLOUD}
            supportedIntegrations={SupportedIntegrations}
          />
          <Typography variant="h6" marginY={2}>
            Compute
          </Typography>
          <AddIntegrations
            user={user}
            category={IntegrationCategories.COMPUTE}
            supportedIntegrations={SupportedIntegrations}
          />
          <Typography variant="h6" marginY={2}>
            Data
          </Typography>
          <AddIntegrations
            user={user}
            category={IntegrationCategories.DATA}
            supportedIntegrations={SupportedIntegrations}
          />
          <Typography variant="h6" marginY={2}>
            Notifications
          </Typography>
          <Typography variant="h6" marginY={2}>
            <AddIntegrations
              user={user}
              category={IntegrationCategories.NOTIFICATION}
              supportedIntegrations={SupportedIntegrations}
            />
          </Typography>
        </Box>

        <Box marginY={3}>
          <Divider />
        </Box>

        <MetadataStorageInfo serverConfig={serverConfig.config} />
        {!isLoading && lastFailedFormattedTimestamp && (
            <Box>
              <Typography variant="body2" fontWeight="fontWeightRegular" marginTop={2} marginBottom={1}>
                The last artifact storage migration, which started at {lastFailedFormattedTimestamp}, has failed!
                As a result, the artifact storage has not changed from `{serverConfig.config?.storageConfig.integration_name}`.
              </Typography>
              <Box
                  sx={{
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

        <Box marginY={3}>
          <Divider />
        </Box>

        <Box>
          <Typography variant="h5" marginY={2}>
            Connected Data Integrations
          </Typography>

          <ConnectedIntegrations user={user} forceLoad={forceLoad} />
        </Box>
      </Box>

      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showDeleteIntegrationSuccessToast}
        key={'workflowheader-delete-success-error-snackbar'}
        autoHideDuration={6000}
        onClose={() => {
          setShowDeleteIntegrationSuccessToast(false);
          location.state = undefined;
        }}
      >
        <Alert severity="success" sx={{ width: '100%' }}>
          {`Successfully deleted ${deleteIntegrationName}`}
        </Alert>
      </Snackbar>
    </Layout>
  );
};

export default IntegrationsPage;
