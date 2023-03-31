import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import Snackbar from '@mui/material/Snackbar';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useLocation, useParams } from 'react-router-dom';

import { DetailIntegrationCard } from '../../../../components/integrations/cards/detailCard';
import AddTableDialog from '../../../../components/integrations/dialogs/addTableDialog';
import DeleteIntegrationDialog from '../../../../components/integrations/dialogs/deleteIntegrationDialog';
import IntegrationDialog from '../../../../components/integrations/dialogs/dialog';
import IntegrationObjectList from '../../../../components/integrations/integrationObjectList';
import OperatorsOnIntegration from '../../../../components/integrations/operatorsOnIntegration';
import DefaultLayout from '../../../../components/layouts/default';
import { BreadcrumbLink } from '../../../../components/layouts/NavBar';
import { handleGetServerConfig } from '../../../../handlers/getServerConfig';
import {
  handleListIntegrationObjects,
  handleLoadIntegrationOperators,
  handleTestConnectIntegration,
  resetEditStatus,
  resetTestConnectStatus,
} from '../../../../reducers/integration';
import { handleLoadIntegrations } from '../../../../reducers/integrations';
import { handleFetchAllWorkflowSummaries } from '../../../../reducers/listWorkflowSummaries';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import {
  IntegrationCategories,
  SupportedIntegrations,
} from '../../../../utils/integrations';
import { isFailed, isLoading, isSucceeded } from '../../../../utils/shared';
import IntegrationOptions from '../../../integrations/options';
import { LayoutProps } from '../../types';

type IntegrationDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const IntegrationDetailsPage: React.FC<IntegrationDetailsPageProps> = ({
  user,
  Layout = DefaultLayout,
}) => {
  const dispatch: AppDispatch = useDispatch();
  const path = useLocation().pathname;

  const integrationId: string = useParams().id;
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [showAddTableDialog, setShowAddTableDialog] = useState(false);
  const [showDeleteTableDialog, setShowDeleteTableDialog] = useState(false);

  const [showTestConnectToast, setShowTestConnectToast] = useState(false);
  const [showConnectSuccessToast, setShowConnectSuccessToast] = useState(false);
  const [showEditSuccessToast, setShowEditSuccessToast] = useState(false);

  const handleCloseConnectSuccessToast = () => {
    setShowConnectSuccessToast(false);
  };

  const handleCloseTestConnectToast = () => {
    setShowTestConnectToast(false);
  };

  const handleCloseEditSuccessToast = () => {
    setShowEditSuccessToast(false);
  };

  const testConnectStatus = useSelector(
    (state: RootState) => state.integrationReducer.testConnectStatus
  );

  const integrations = useSelector(
    (state: RootState) => state.integrationsReducer.integrations
  );

  const isListObjectsLoading = useSelector((state: RootState) =>
    isLoading(state.integrationReducer.objectNames.status)
  );

  const serverConfig = useSelector(
    (state: RootState) => state.serverConfigReducer
  );

  const selectedIntegration = integrations[integrationId];

  // Using the ListIntegrationsRoute.
  // ENG-1036: We should create a route where we can pass in the integrationId and get the associated metadata and switch to using that.
  useEffect(() => {
    dispatch(handleLoadIntegrations({ apiKey: user.apiKey }));
    dispatch(handleFetchAllWorkflowSummaries({ apiKey: user.apiKey }));
    dispatch(
      handleLoadIntegrationOperators({
        apiKey: user.apiKey,
        integrationId: integrationId,
      })
    );
  }, [dispatch, integrationId, user.apiKey]);

  useEffect(() => {
    if (!isLoading(testConnectStatus)) {
      setShowTestConnectToast(false);
    }

    if (isSucceeded(testConnectStatus)) {
      setShowConnectSuccessToast(true);
      dispatch(resetTestConnectStatus());
    }
  }, [dispatch, testConnectStatus]);

  useEffect(() => {
    if (selectedIntegration && selectedIntegration.name) {
      document.title = `Integration Details: ${selectedIntegration.name} | Aqueduct`;
    } else {
      document.title = `Integration Details | Aqueduct`;
    }

    if (
      selectedIntegration &&
      SupportedIntegrations[selectedIntegration.service].category ===
        IntegrationCategories.DATA
    ) {
      dispatch(
        handleListIntegrationObjects({
          apiKey: user.apiKey,
          integrationId: integrationId,
        })
      );
    }
  }, [selectedIntegration]);

  // Disable deletion of a storage integration.
  useEffect(() => {
    async function fetchServerConfig() {
      if (user) {
        await dispatch(handleGetServerConfig({ apiKey: user.apiKey }));
      }
    }

    fetchServerConfig();
  }, [user.apiKey]);

  if (!integrations || !selectedIntegration) {
    return null;
  }

  return (
    <Layout
      breadcrumbs={[
        BreadcrumbLink.HOME,
        BreadcrumbLink.INTEGRATIONS,
        new BreadcrumbLink(path, selectedIntegration.name),
      ]}
      user={user}
    >
      <Box sx={{ paddingBottom: '4px' }}>
        <Box display="flex" flexDirection="row" alignContent="top">
          <DetailIntegrationCard
            integration={selectedIntegration}
            connectStatus={testConnectStatus}
          />
          <IntegrationOptions
            integration={selectedIntegration}
            onUploadCsv={() => setShowAddTableDialog(true)}
            onTestConnection={() => {
              dispatch(
                handleTestConnectIntegration({
                  apiKey: user.apiKey,
                  integrationId: selectedIntegration.id,
                })
              );
              setShowTestConnectToast(true);
            }}
            onEdit={() => setShowEditDialog(true)}
            onDeleteIntegration={() => {
              setShowDeleteTableDialog(true);
            }}
            allowDeletion={
              serverConfig.config?.storageConfig.integration_name !==
              selectedIntegration.name
            }
          />
        </Box>

        {showDeleteTableDialog && (
          <DeleteIntegrationDialog
            user={user}
            integrationId={selectedIntegration.id}
            integrationName={selectedIntegration.name}
            integrationType={selectedIntegration.service}
            config={selectedIntegration.config}
            onCloseDialog={() => setShowDeleteTableDialog(false)}
          />
        )}

        {testConnectStatus && isFailed(testConnectStatus) && (
          <Alert severity="error" sx={{ marginTop: 2 }}>
            Test-connect failed with error:
            <br></br>
            <pre>{testConnectStatus.err}</pre>
          </Alert>
        )}

        {selectedIntegration.name === 'aqueduct_demo' && (
          <Typography variant="body1" sx={{ my: 1 }}>
            You can see the documentation for the Aqueduct Demo database{' '}
            <Link href="https://docs.aqueducthq.com/integrations/aqueduct-demo-integration">
              here
            </Link>
            .
          </Typography>
        )}

        {SupportedIntegrations[selectedIntegration.service].category ===
          IntegrationCategories.DATA && (
          <IntegrationObjectList
            user={user}
            integration={selectedIntegration}
          />
        )}

        {SupportedIntegrations[selectedIntegration.service].category !==
          IntegrationCategories.NOTIFICATION && (
          <Box sx={{ mt: 4 }}>
            <Typography
              variant="h5"
              gutterBottom
              component="div"
              sx={{ mb: 4 }}
            >
              Workflows
            </Typography>
            <OperatorsOnIntegration />
          </Box>
        )}
      </Box>

      {showAddTableDialog && (
        <AddTableDialog
          user={user}
          integrationId={selectedIntegration.id}
          onCloseDialog={() => setShowAddTableDialog(false)}
          onConnect={() => {
            if (!isListObjectsLoading) {
              dispatch(
                handleListIntegrationObjects({
                  apiKey: user.apiKey,
                  integrationId: integrationId,
                  forceLoad: true,
                })
              );
            }

            setShowAddTableDialog(false);
          }}
        />
      )}

      {showEditDialog && (
        <IntegrationDialog
          user={user}
          service={selectedIntegration.service}
          onSuccess={() => setShowEditSuccessToast(true)}
          onCloseDialog={() => {
            setShowEditDialog(false);
            dispatch(resetEditStatus());
          }}
          integrationToEdit={selectedIntegration}
        />
      )}

      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showTestConnectToast}
        onClose={handleCloseTestConnectToast}
        key={'integration-test-connect-snackbar'}
        autoHideDuration={6000}
      >
        <Alert
          onClose={handleCloseTestConnectToast}
          severity="info"
          sx={{ width: '100%' }}
        >
          {`Attempting to connect to ${selectedIntegration.name}`}
        </Alert>
      </Snackbar>

      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showConnectSuccessToast}
        onClose={handleCloseConnectSuccessToast}
        key={'integration-connect-success-snackbar'}
        autoHideDuration={6000}
      >
        <Alert
          onClose={handleCloseConnectSuccessToast}
          severity="success"
          sx={{ width: '100%' }}
        >
          {`Successfully connected to ${selectedIntegration.name}`}
        </Alert>
      </Snackbar>

      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showEditSuccessToast}
        onClose={handleCloseEditSuccessToast}
        key={'integration-edit-success-snackbar'}
        autoHideDuration={6000}
      >
        <Alert
          onClose={handleCloseEditSuccessToast}
          severity="success"
          sx={{ width: '100%' }}
        >
          {`Successfully updated ${selectedIntegration.name}`}
        </Alert>
      </Snackbar>
    </Layout>
  );
};

export default IntegrationDetailsPage;
