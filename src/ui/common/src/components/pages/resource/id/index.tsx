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

import DefaultLayout from '../../../../components/layouts/default';
import { BreadcrumbLink } from '../../../../components/layouts/NavBar';
import AddTableDialog from '../../../../components/resources/dialogs/addTableDialog';
import DeleteResourceDialog from '../../../../components/resources/dialogs/deleteResourceDialog';
import ResourceDialog from '../../../../components/resources/dialogs/dialog';
import ResourceObjectList from '../../../../components/resources/resourceObjectList';
import {
  useResourceOperatorsGetQuery,
  useResourceWorkflowsGetQuery,
} from '../../../../handlers/AqueductApi';
import { handleGetServerConfig } from '../../../../handlers/getServerConfig';
import { OperatorResponse } from '../../../../handlers/responses/node';
import { handleFetchAllWorkflowSummaries } from '../../../../reducers/listWorkflowSummaries';
import {
  handleListResourceObjects,
  handleTestConnectResource,
  resetEditStatus,
  resetTestConnectStatus,
} from '../../../../reducers/resource';
import { handleLoadResources } from '../../../../reducers/resources';
import { AppDispatch, RootState } from '../../../../stores/store';
import { theme } from '../../../../styles/theme/theme';
import UserProfile from '../../../../utils/auth';
import {
  isNotificationResource,
  ResourceCategories,
  resourceExecState,
} from '../../../../utils/resources';
import ExecutionStatus, {
  isFailed,
  isLoading,
  isSucceeded,
} from '../../../../utils/shared';
import SupportedResources from '../../../../utils/SupportedResources';
import { ResourceHeaderDetailsCard } from '../../../resources/cards/headerDetailsCard';
import { ResourceFieldsDetailsCard } from '../../../resources/cards/resourceFieldsDetailsCard';
import { ErrorSnackbar } from '../../../resources/errorSnackbar';
import { getNumWorkflowsUsingMessage } from '../../../resources/numWorkflowsUsingMsg';
import ResourceOptions, {
  ResourceOptionsButtonWidth,
} from '../../../resources/options';
import ResourceWorkflowSummaryCards from '../../../resources/resourceWorkflowSummaryCards';
import { LayoutProps } from '../../types';

type ResourceDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const ResourceDetailsPage: React.FC<ResourceDetailsPageProps> = ({
  user,
  Layout = DefaultLayout,
}) => {
  const dispatch: AppDispatch = useDispatch();
  const path = useLocation().pathname;

  const resourceId: string = useParams().id;
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
    (state: RootState) => state.resourceReducer.testConnectStatus
  );

  const resources = useSelector(
    (state: RootState) => state.resourcesReducer.resources
  );

  const isListObjectsLoading = useSelector((state: RootState) =>
    isLoading(state.resourceReducer.objectNames.status)
  );

  const serverConfig = useSelector(
    (state: RootState) => state.serverConfigReducer
  );

  const selectedResource = resources[resourceId];
  const resourceClass = SupportedResources[selectedResource?.service];

  // Using the ListResourcesRoute.
  // ENG-1036: We should create a route where we can pass in the resourceId and get the associated metadata and switch to using that.
  useEffect(() => {
    dispatch(handleLoadResources({ apiKey: user.apiKey }));
    dispatch(handleFetchAllWorkflowSummaries({ apiKey: user.apiKey }));
  }, [dispatch, resourceId, user.apiKey]);

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
    if (selectedResource && selectedResource.name) {
      document.title = `Resource Details: ${selectedResource.name} | Aqueduct`;
    } else {
      document.title = `Resource Details | Aqueduct`;
    }

    if (
      selectedResource &&
      SupportedResources[selectedResource.service].category ===
        ResourceCategories.DATA
    ) {
      dispatch(
        handleListResourceObjects({
          apiKey: user.apiKey,
          resourceId: resourceId,
        })
      );
    }
  }, [dispatch, selectedResource, resourceId, user]);

  // Load the server config to check if the selected resource is currently being used as storage.
  // If that is the case, we hide the option to delete the resource from the user.
  useEffect(() => {
    async function fetchServerConfig() {
      if (user) {
        await dispatch(handleGetServerConfig({ apiKey: user.apiKey }));
      }
    }

    fetchServerConfig();
  }, [dispatch, user]);

  const {
    data: workflowAndDagIDs,
    error: fetchWorkflowsError,

    // Needed to rename this since we're importing an `isLoading` is that was causing problems.
    isLoading: fetchWorkflowsIsLoading,
  } = useResourceWorkflowsGetQuery({
    apiKey: user.apiKey,
    resourceId: resourceId,
  });

  const { data: resourceOperators } = useResourceOperatorsGetQuery({
    apiKey: user.apiKey,
    resourceId: resourceId,
  });

  // Using the latest `dag_id` as the common key to bind the workflow to its latest operators.
  const workflowIDToLatestOperators: {
    [workflowID: string]: OperatorResponse[];
  } = {};
  if (workflowAndDagIDs && resourceOperators) {
    // Reorganize the operators to be keyed by their `dag_id`.
    const operatorsByDagID: { [dagID: string]: OperatorResponse[] } = {};
    resourceOperators.forEach((operator) => {
      if (operatorsByDagID[operator.dag_id]) {
        operatorsByDagID[operator.dag_id].push(operator);
      } else {
        operatorsByDagID[operator.dag_id] = [operator];
      }
    });

    workflowAndDagIDs.forEach((workflowAndDagID) => {
      // If we're displaying a notification, there won't be only operators, but we
      // want to include the workflows.
      if (isNotificationResource(selectedResource)) {
        workflowIDToLatestOperators[workflowAndDagID.id] = [];
      } else if (operatorsByDagID[workflowAndDagID.dag_id]) {
        workflowIDToLatestOperators[workflowAndDagID.id] =
          operatorsByDagID[workflowAndDagID.dag_id];
      }
    });
  }

  if (fetchWorkflowsIsLoading || !selectedResource || !resourceClass) {
    return null;
  }

  // We only count workflows if their latest run has used this resource.
  let numWorkflowsUsingMsg = '';
  if (!fetchWorkflowsError && workflowAndDagIDs) {
    numWorkflowsUsingMsg = getNumWorkflowsUsingMessage(
      Object.keys(workflowIDToLatestOperators).length
    );
  }

  if (!resources || !selectedResource) {
    return null;
  }

  const selectedResourceExecState = resourceExecState(selectedResource);
  return (
    <Layout
      breadcrumbs={[
        BreadcrumbLink.HOME,
        BreadcrumbLink.RESOURCES,
        new BreadcrumbLink(path, selectedResource.name),
      ]}
      user={user}
    >
      <ErrorSnackbar
        shouldShow={fetchWorkflowsError !== undefined}
        errMsg={
          'Unexpected error occurred when fetching workflows associated with this resource. Please try again.'
        }
      />

      <Box sx={{ paddingBottom: '4px' }}>
        <Box display="flex" flexDirection="row" alignContent="top">
          <Box
            sx={{
              flex: 1,
              width: `calc(100% - ${ResourceOptionsButtonWidth})`,
            }}
          >
            <Box display="flex" flexDirection="row" alignContent="bottom">
              <ResourceHeaderDetailsCard
                resource={selectedResource}
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

          <ResourceOptions
            resource={selectedResource}
            onUploadCsv={() => setShowAddTableDialog(true)}
            onTestConnection={() => {
              dispatch(
                handleTestConnectResource({
                  apiKey: user.apiKey,
                  resourceId: selectedResource.id,
                })
              );
              setShowTestConnectToast(true);
            }}
            onEdit={() => setShowEditDialog(true)}
            onDeleteResource={() => {
              setShowDeleteTableDialog(true);
            }}
            allowDeletion={
              serverConfig.config?.storageConfig.resource_name !==
              selectedResource.name
            }
          />
        </Box>

        {selectedResourceExecState.status === ExecutionStatus.Failed && (
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
              {`${selectedResourceExecState.error.tip}\n\n${selectedResourceExecState?.error.context}`}
            </Typography>
          </Box>
        )}

        {serverConfig.config?.storageConfig.resource_name ===
          selectedResource.name && (
          <Alert severity="info" sx={{ marginTop: 2 }}>
            This resource cannot be deleted because it is currently being used
            as artifact storage. To delete this resource, please migrate your
            artifact storage elsewhere first.
          </Alert>
        )}

        {showDeleteTableDialog && (
          <DeleteResourceDialog
            user={user}
            resourceId={selectedResource.id}
            resourceName={selectedResource.name}
            resourceType={selectedResource.service}
            config={selectedResource.config}
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

        {selectedResource.name === 'Demo' &&
          selectedResource.service == 'SQLite' && (
            <Typography variant="body1" sx={{ my: 1 }}>
              You can see the documentation for the Aqueduct Demo database{' '}
              <Link href="https://docs.aqueducthq.com/resources/aqueduct-demo-resource">
                here
              </Link>
              .
            </Typography>
          )}

        {showResourceDetails && (
          <Box sx={{ my: 1, mt: 2 }}>
            <ResourceFieldsDetailsCard
              resource={selectedResource}
              detailedView={true}
            />
          </Box>
        )}

        {SupportedResources[selectedResource.service].category ===
          ResourceCategories.DATA && (
          <ResourceObjectList user={user} resource={selectedResource} />
        )}

        <Box sx={{ mt: 4 }}>
          <Typography variant="h5" gutterBottom component="div" sx={{ mb: 4 }}>
            Workflows
          </Typography>

          <ResourceWorkflowSummaryCards
            resource={selectedResource}
            workflowIDToLatestOperators={workflowIDToLatestOperators}
          />
        </Box>
      </Box>

      {showAddTableDialog && (
        <AddTableDialog
          user={user}
          resourceId={selectedResource.id}
          onCloseDialog={() => setShowAddTableDialog(false)}
          onConnect={() => {
            if (!isListObjectsLoading) {
              dispatch(
                handleListResourceObjects({
                  apiKey: user.apiKey,
                  resourceId: resourceId,
                  forceLoad: true,
                })
              );
            }

            setShowAddTableDialog(false);
          }}
        />
      )}

      {showEditDialog && (
        <ResourceDialog
          user={user}
          service={selectedResource.service}
          onSuccess={() => setShowEditSuccessToast(true)}
          onCloseDialog={() => {
            setShowEditDialog(false);
            dispatch(resetEditStatus());
          }}
          resourceToEdit={selectedResource}
          dialogContent={resourceClass.dialog}
          validationSchema={resourceClass.validationSchema(!!selectedResource)}
        />
      )}

      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showTestConnectToast}
        onClose={handleCloseTestConnectToast}
        key={'resource-test-connect-snackbar'}
        autoHideDuration={6000}
      >
        <Alert
          onClose={handleCloseTestConnectToast}
          severity="info"
          sx={{ width: '100%' }}
        >
          {`Attempting to connect to ${selectedResource.name}`}
        </Alert>
      </Snackbar>

      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showConnectSuccessToast}
        onClose={handleCloseConnectSuccessToast}
        key={'resource-connect-success-snackbar'}
        autoHideDuration={6000}
      >
        <Alert
          onClose={handleCloseConnectSuccessToast}
          severity="success"
          sx={{ width: '100%' }}
        >
          {`Successfully connected to ${selectedResource.name}`}
        </Alert>
      </Snackbar>

      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showEditSuccessToast}
        onClose={handleCloseEditSuccessToast}
        key={'resource-edit-success-snackbar'}
        autoHideDuration={6000}
      >
        <Alert
          onClose={handleCloseEditSuccessToast}
          severity="success"
          sx={{ width: '100%' }}
        >
          {`Successfully updated ${selectedResource.name}`}
        </Alert>
      </Snackbar>
    </Layout>
  );
};

export default ResourceDetailsPage;
