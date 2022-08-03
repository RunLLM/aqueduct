import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useParams } from 'react-router-dom';

import { DetailIntegrationCard } from '../../../../components/integrations/cards/detailCard';
import AddTableDialog from '../../../../components/integrations/dialogs/addTableDialog';
import TestConnectDialog from '../../../../components/integrations/dialogs/testConnectDialog';
import IntegrationObjectList from '../../../../components/integrations/integrationObjectList';
import OperatorsOnIntegration from '../../../../components/integrations/operatorsOnIntegration';
import DefaultLayout from '../../../../components/layouts/default';
import {
  handleListIntegrationObjects,
  handleLoadIntegrationOperators,
  handleTestConnectIntegration,
} from '../../../../reducers/integration';
import { handleLoadIntegrations } from '../../../../reducers/integrations';
import { handleFetchAllWorkflowSummaries } from '../../../../reducers/listWorkflowSummaries';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import { Integration } from '../../../../utils/integrations';
import { isLoading } from '../../../../utils/shared';
import IntegrationButtonGroup from '../../../integrations/buttonGroup';
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
  const integrationId: string = useParams().id;
  const [showAddTableDialog, setShowAddTableDialog] = useState(false);
  const [showTestConnectDialog, setShowTestConnectDialog] = useState(false);

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

  const integrations = useSelector(
    (state: RootState) => state.integrationsReducer.integrations
  );

  const isListObjectsLoading = useSelector((state: RootState) =>
    isLoading(state.integrationReducer.objectNames.status)
  );

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
          <DetailIntegrationCard integration={selectedIntegration} />
          <IntegrationButtonGroup
            integration={selectedIntegration}
            onUploadCsv={() => setShowAddTableDialog(true)}
            onTestConnection={() => {
              dispatch(
                handleTestConnectIntegration({
                  apiKey: user.apiKey,
                  integrationId: selectedIntegration.id,
                })
              );
              setShowTestConnectDialog(true);
            }}
          />
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
        {showTestConnectDialog && (
          <TestConnectDialog
            onCloseDialog={() => setShowTestConnectDialog(false)}
          />
        )}
        <IntegrationObjectList user={user} integration={selectedIntegration} />
        <Typography
          variant="h4"
          gutterBottom
          component="div"
          sx={{ marginY: 4 }}
        >
          Workflows
        </Typography>
        <OperatorsOnIntegration />
      </Box>
    </Layout>
  );
};

export default IntegrationDetailsPage;
