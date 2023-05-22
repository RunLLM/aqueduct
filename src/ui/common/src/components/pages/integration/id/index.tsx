import { faEllipsis } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Tooltip } from '@mui/material';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import Snackbar from '@mui/material/Snackbar';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useLocation, useParams } from 'react-router-dom';

import AddTableDialog from '../../../../components/integrations/dialogs/addTableDialog';
import DeleteIntegrationDialog from '../../../../components/integrations/dialogs/deleteIntegrationDialog';
import IntegrationDialog from '../../../../components/integrations/dialogs/dialog';
import IntegrationObjectList from '../../../../components/integrations/integrationObjectList';
import DefaultLayout from '../../../../components/layouts/default';
import { BreadcrumbLink } from '../../../../components/layouts/NavBar';
import {
  useIntegrationOperatorsGetQuery,
  useIntegrationWorkflowsGetQuery,
} from '../../../../handlers/AqueductApi';
import { handleGetServerConfig } from '../../../../handlers/getServerConfig';
import { OperatorResponse } from '../../../../handlers/responses/node';
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
import { theme } from '../../../../styles/theme/theme';
import UserProfile from '../../../../utils/auth';
import {
  IntegrationCategories,
  isNotificationIntegration,
  resourceExecState,
  SupportedIntegrations,
} from '../../../../utils/integrations';
import ExecutionStatus, {
  isFailed,
  isLoading,
  isSucceeded,
} from '../../../../utils/shared';
import SupportedIntegrations from '../../../../utils/SupportedIntegrations';
import { ResourceHeaderDetailsCard } from '../../../integrations/cards/headerDetailsCard';
import { ResourceFieldsDetailsCard } from '../../../integrations/cards/resourceFieldsDetailsCard';
import { ErrorSnackbar } from '../../../integrations/errorSnackbar';
import IntegrationWorkflowSummaryCards from '../../../integrations/integrationWorkflowSummaryCards';
import { getNumWorkflowsUsingMessage } from '../../../integrations/numWorkflowsUsingMsg';
import IntegrationOptions, {
  IntegrationOptionsButtonWidth,
} from '../../../integrations/options';
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
  const [showResourceDetails, setShowResourceDetails] = useState(false);

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

  // Load the server config to check if the selected integration is currently being used as storage.
  // If that is the case, we hide the option to delete the integration from the user.
  useEffect(() => {
    async function fetchServerConfig() {
      if (user) {
        await dispatch(handleGetServerConfig({ apiKey: user.apiKey }));
      }
    }

    fetchServerConfig();
  }, [user.apiKey]);

  const {
    data: workflowAndDagIDs,
    error: fetchWorkflowsError,

    // Needed to rename this since we're importing an `isLoading` is that was causing problems.
    isLoading: fetchWorkflowsIsLoading,
  } = useIntegrationWorkflowsGetQuery({
    apiKey: user.apiKey,
    integrationId: integrationId,
  });

  const {
    data: integrationOperators,
    error: testOpsErr,
    isLoading: testOpsIsLoading,
  } = useIntegrationOperatorsGetQuery({
    apiKey: user.apiKey,
    integrationId: integrationId,
  });

  // Using the latest `dag_id` as the common key to bind the workflow to its latest operators.
  const workflowIDToLatestOperators: {
    [workflowID: string]: OperatorResponse[];
  } = {};
  if (workflowAndDagIDs && integrationOperators) {
    // Reorganize the operators to be keyed by their `dag_id`.
    const operatorsByDagID: { [dagID: string]: OperatorResponse[] } = {};
    integrationOperators.forEach((operator) => {
      if (operatorsByDagID[operator.dag_id]) {
        operatorsByDagID[operator.dag_id].push(operator);
      } else {
        operatorsByDagID[operator.dag_id] = [operator];
      }
    });

    workflowAndDagIDs.forEach((workflowAndDagID) => {
      // If we're displaying a notification, there won't be only operators, but we
      // want to include the workflows.
      if (isNotificationIntegration(selectedIntegration)) {
        workflowIDToLatestOperators[workflowAndDagID.id] = [];
      } else if (operatorsByDagID[workflowAndDagID.dag_id]) {
        workflowIDToLatestOperators[workflowAndDagID.id] =
          operatorsByDagID[workflowAndDagID.dag_id];
      }
    });
  }

  if (fetchWorkflowsIsLoading) {
    return null;
  }

  // We only count workflows if their latest run has used this resource.
  let numWorkflowsUsingMsg = '';
  if (!fetchWorkflowsError && workflowAndDagIDs) {
    numWorkflowsUsingMsg = getNumWorkflowsUsingMessage(
      Object.keys(workflowIDToLatestOperators).length
    );
  }

  if (!integrations || !selectedIntegration) {
    return null;
  }

  const selectedIntegrationExecState = resourceExecState(selectedIntegration);

  console.log('selectedIntegration: ', selectedIntegration);

  return (
    <Layout
      breadcrumbs={[
        BreadcrumbLink.HOME,
        BreadcrumbLink.INTEGRATIONS,
        new BreadcrumbLink(path, selectedIntegration.name),
      ]}
      user={user}
    >
      <ErrorSnackbar
        shouldShow={fetchWorkflowsError !== undefined}
        errMsg={
          'Unexpected error occurred when fetching workflows associated with this integration. Please try again.'
        }
      />

      <Box sx={{ paddingBottom: '4px' }}>
        <Box display="flex" flexDirection="row" alignContent="top">
          <Box
            sx={{
              flex: 1,
              width: `calc(100% - ${IntegrationOptionsButtonWidth})`,
            }}
          >
            <Box display="flex" flexDirection="row" alignContent="bottom">
              <ResourceHeaderDetailsCard
                integration={selectedIntegration}
                numWorkflowsUsingMsg={numWorkflowsUsingMsg}
              />

              <Box
                sx={{
                  fontSize: '16px',
                  p: 1,
                  ml: 1,
                  height: '32px',
                  borderRadius: '8px',
                  ':hover': {
                    backgroundColor: theme.palette.gray[50],
                  },
                  cursor: 'pointer',
                }}
                onClick={() => setShowResourceDetails(!showResourceDetails)}
              >
                <Tooltip title="See more" arrow>
                  <FontAwesomeIcon
                    icon={faEllipsis}
                    style={{
                      transition: 'transform 200ms',
                    }}
                  />
                </Tooltip>
              </Box>
            </Box>
          </Box>

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
            onEdit={() => {
              console.log('inside onEdit()');
              setShowEditDialog(true)
            }}
            onDeleteIntegration={() => {
              setShowDeleteTableDialog(true);
            }}
            allowDeletion={
              serverConfig.config?.storageConfig.integration_name !==
              selectedIntegration.name
            }
          />
        </Box>

        {selectedIntegrationExecState.status === ExecutionStatus.Failed && (
          <Box
            sx={{
              backgroundColor: theme.palette.red[100],
              borderRadius: 2,
              color: theme.palette.red[600],
              p: 2,
              height: 'fit-content',
              width: '100%',
              my: 1,
            }}
          >
            <Typography variant="body2" style={{ whiteSpace: 'pre-wrap' }}>
              {`${selectedIntegrationExecState.error.tip}\n\n${selectedIntegrationExecState?.error.context}`}
            </Typography>
          </Box>
        )}

        {serverConfig.config?.storageConfig.integration_name ===
          selectedIntegration.name && (
          <Alert severity="info" sx={{ marginTop: 2 }}>
            This integration cannot be deleted because it is currently being
            used as artifact storage. To delete this integration, please migrate
            your artifact storage elsewhere first.
          </Alert>
        )}

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

        {selectedIntegration.name === 'Demo' &&
          selectedIntegration.service == 'SQLite' && (
            <Typography variant="body1" sx={{ my: 1 }}>
              You can see the documentation for the Aqueduct Demo database{' '}
              <Link href="https://docs.aqueducthq.com/integrations/aqueduct-demo-integration">
                here
              </Link>
              .
            </Typography>
          )}

        {showResourceDetails && (
          <Box sx={{ my: 1, mt: 2 }}>
            <ResourceFieldsDetailsCard
              integration={selectedIntegration}
              detailedView={true}
            />
          </Box>
        )}

        {SupportedIntegrations[selectedIntegration.service].category ===
          IntegrationCategories.DATA && (
          <IntegrationObjectList
            user={user}
            integration={selectedIntegration}
          />
        )}

        <Box sx={{ mt: 4 }}>
          <Typography variant="h5" gutterBottom component="div" sx={{ mb: 4 }}>
            Workflows
          </Typography>

          <IntegrationWorkflowSummaryCards
            integration={selectedIntegration}
            workflowIDToLatestOperators={workflowIDToLatestOperators}
          />
        </Box>
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

      {/* TODO: Get the selectedIntegration from the map of integrations. Then figure out which dialog that we should be rendering. Then pass in the information.
          Doing it this way is causing an infinite loop
      */}
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
