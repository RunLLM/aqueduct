import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Snackbar from '@mui/material/Snackbar';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate, useParams } from 'react-router-dom';

import { DetailIntegrationCard } from '../../../../components/integrations/cards/detailCard';
import AddTableDialog from '../../../../components/integrations/dialogs/addTableDialog';
import IntegrationObjectList from '../../../../components/integrations/integrationObjectList';
import OperatorsOnIntegration from '../../../../components/integrations/operatorsOnIntegration';
import DefaultLayout from '../../../../components/layouts/default';
import {
  handleDeleteIntegration,
  handleListIntegrationObjects,
  handleLoadIntegrationOperators,
  handleTestConnectIntegration,
} from '../../../../reducers/integration';
import { handleLoadIntegrations } from '../../../../reducers/integrations';
import { handleFetchAllWorkflowSummaries } from '../../../../reducers/listWorkflowSummaries';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import { Integration } from '../../../../utils/integrations';
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
  const navigate = useNavigate();
  const integrationId: string = useParams().id;
  const [showAddTableDialog, setShowAddTableDialog] = useState(false);

  const [showTestConnectToast, setShowTestConnectToast] = useState(false);
  const [showDeleteIntegrationToast, setShowDeleteIntegrationToast] =
    useState(false);
  const [showConnectSuccessToast, setShowConnectSuccessToast] = useState(false);
  const [showDeleteSuccessToast, setShowDeleteSuccessToast] = useState(false);

  const handleCloseConnectSuccessToast = () => {
    setShowConnectSuccessToast(false);
  };

  const handleCloseDeleteSuccessToast = () => {
    setShowDeleteSuccessToast(false);
  };

  const handleCloseTestConnectToast = () => {
    setShowTestConnectToast(false);
  };

  const handleCloseDeleteIntegrationToast = () => {
    setShowDeleteIntegrationToast(false);
  };

  const testConnectStatus = useSelector(
    (state: RootState) => state.integrationReducer.connectionStatus
  );

  const deleteIntegrationStatus = useSelector(
    (state: RootState) => state.integrationReducer.deletionStatus
  );

  const integrations = useSelector(
    (state: RootState) => state.integrationsReducer.integrations
  );

  const isListObjectsLoading = useSelector((state: RootState) =>
    isLoading(state.integrationReducer.objectNames.status)
  );

  // Using the ListIntegrationsRoute.
  // ENG-1036: We should create a route where we can pass in the integrationId and get the associated metadata and switch to using that.
  useEffect(() => {
    dispatch(handleLoadIntegrations({ apiKey: user.apiKey }));
    dispatch(
      handleListIntegrationObjects({
        apiKey: user.apiKey,
        integrationId: integrationId,
      })
    );
    dispatch(
      handleLoadIntegrationOperators({
        apiKey: user.apiKey,
        integrationId: integrationId,
      })
    );
    dispatch(handleFetchAllWorkflowSummaries({ apiKey: user.apiKey }));
  }, []);

  useEffect(() => {
    if (!isLoading(testConnectStatus)) {
      setShowTestConnectToast(false);
    }

    if (isSucceeded(testConnectStatus)) {
      setShowConnectSuccessToast(true);
    }
  }, [testConnectStatus]);

  useEffect(() => {
    if (!isLoading(deleteIntegrationStatus)) {
      setShowDeleteIntegrationToast(false);
    }

    if (isSucceeded(deleteIntegrationStatus)) {
      setShowDeleteSuccessToast(true);
      // Reload integrations because deleted
      dispatch(
        handleLoadIntegrations({ apiKey: user.apiKey, forceLoad: true })
      );
      // Integration no longer exists. Navigate back to integration list after a short pause.
      setTimeout(() => navigate('/integrations'), 2000);
    }
  }, [deleteIntegrationStatus]);

  let selectedIntegration = null;

  if (integrations) {
    (integrations as Integration[]).forEach((integration) => {
      if (integration.id === integrationId) {
        selectedIntegration = integration;
      }
    });
  }

  useEffect(() => {
    if (selectedIntegration && selectedIntegration.name) {
      document.title = `Integration Details: ${selectedIntegration.name} | Aqueduct`;
    } else {
      document.title = `Integration Details | Aqueduct`;
    }
  }, []);

  if (!integrations || !selectedIntegration) {
    return null;
  }

  return (
    <Layout user={user}>
      <Box sx={{ paddingBottom: '4px' }}>
        <Typography variant="h2" gutterBottom component="div">
          Integration Details
        </Typography>
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
            onDeleteIntegration={() => {
              dispatch(
                handleDeleteIntegration({
                  apiKey: user.apiKey,
                  integrationId: selectedIntegration.id,
                })
              );
              setShowDeleteIntegrationToast(true);
            }}
          />
        </Box>
        {testConnectStatus && isFailed(testConnectStatus) && (
          <Alert severity="error" sx={{ marginTop: 2 }}>
            Test-connect failed with error:
            <br></br>
            <pre>{testConnectStatus.err}</pre>
          </Alert>
        )}
        {deleteIntegrationStatus && isFailed(deleteIntegrationStatus) && (
          <Alert severity="error" sx={{ marginTop: 2 }}>
            Integration deletion failed with error:
            <br></br>
            <pre>{deleteIntegrationStatus.err}</pre>
          </Alert>
        )}
        <IntegrationObjectList user={user} integration={selectedIntegration} />
        <Typography
          variant="h4"
          gutterBottom
          component="div"
          sx={{ marginY: 4, mt: 4 }}
        >
          Workflows
        </Typography>
        <OperatorsOnIntegration />
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
      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showTestConnectToast}
        onClose={handleCloseTestConnectToast}
        key={'workflowheader-connect-attempt-info-snackbar'}
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
        open={showDeleteIntegrationToast}
        onClose={handleCloseDeleteIntegrationToast}
        key={'workflowheader-delete-attempt-info-snackbar'}
        autoHideDuration={6000}
      >
        <Alert
          onClose={handleCloseDeleteIntegrationToast}
          severity="info"
          sx={{ width: '100%' }}
        >
          {`Attempting to delete ${selectedIntegration.name}`}
        </Alert>
      </Snackbar>
      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showConnectSuccessToast}
        onClose={handleCloseConnectSuccessToast}
        key={'workflowheader-connect-success-error-snackbar'}
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
        open={showDeleteSuccessToast}
        onClose={handleCloseDeleteSuccessToast}
        key={'workflowheader-delete-success-error-snackbar'}
        autoHideDuration={6000}
      >
        <Alert
          onClose={handleCloseDeleteSuccessToast}
          severity="success"
          sx={{ width: '100%' }}
        >
          {`Successfully deleted ${selectedIntegration.name}`}
        </Alert>
      </Snackbar>
    </Layout>
  );
};

export default IntegrationDetailsPage;
