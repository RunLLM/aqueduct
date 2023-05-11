import { Typography } from '@mui/material';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Snackbar from '@mui/material/Snackbar';
import React, { useState } from 'react';
import {useDispatch, useSelector} from 'react-redux';

import { resetConnectNewStatus } from '../../reducers/integration';
import {AppDispatch, RootState} from '../../stores/store';
import { theme } from '../../styles/theme/theme';
import UserProfile from '../../utils/auth';
import { Info, IntegrationCategories, Service, ServiceInfoMap, SupportedIntegrations } from '../../utils/integrations';
import IntegrationDialog from './dialogs/dialog';
import IntegrationLogo from './logo';

type Props = {
  user: UserProfile;
  supportedIntegrations: ServiceInfoMap;
  category: string;
};

const AddIntegrations: React.FC<Props> = ({
  user,
  supportedIntegrations,
  category,
}) => {
  const [showSuccessToast, setShowSuccessToast] = useState<Service>(null);
  const handleSuccessToastClose = () => {
    setShowSuccessToast(null);
  };
  const [showMigrationDialog, setShowMigrationDialog] = useState(false);

  return (
    <Box>
      {showMigrationDialog && (
        <Alert
          onClose={() => {
            setShowMigrationDialog(false);
          }}
          severity="info"
          sx={{ margin: 1 }}
        >
          {`Storage migration is in progress. The server will be temporarily unavailable. Please refresh the page to check if the server is ready.`}
        </Alert>
      )}
      <Box sx={{ width: '100%', display: 'flex', flexWrap: 'wrap' }}>
        {Object.entries(supportedIntegrations)
          .filter(([svc]) => svc !== 'Aqueduct')
          .sort(([name1], [name2]) => name1.localeCompare(name2))
          .map(([svc, integration]) => {
            return (
              <AddIntegrationListItem
                key={svc as string}
                svc={svc}
                dialog={integration.dialog}
                integration={integration}
                category={category}
                handleSuccessToastClose={handleSuccessToastClose}
                user={user}
                showSuccessToast={showSuccessToast}
                setShowSuccessToast={setShowSuccessToast}
                setShowMigrationDialog={setShowMigrationDialog}
              />
            );
          })}
      </Box>
    </Box>
  );
};

interface AddIntegrationListItemProps {
  svc: string;
  integration: Info;
  category: string;
  user: UserProfile;
  showSuccessToast: string;
  // callback functions
  handleSuccessToastClose: () => void;
  setShowSuccessToast: React.Dispatch<React.SetStateAction<Service>>;
  setShowMigrationDialog: React.Dispatch<React.SetStateAction<boolean>>;
  dialog: React.FC;
}

const AddIntegrationListItem: React.FC<AddIntegrationListItemProps> = ({
  svc,
  integration,
  category,
  user,
  setShowMigrationDialog,
  handleSuccessToastClose,
  showSuccessToast,
  setShowSuccessToast,
  dialog,
}) => {
  const dispatch: AppDispatch = useDispatch();
  const service = svc as Service;
  const [showDialog, setShowDialog] = useState(false);

  // If this is a conda integration, check if it has already been registered.
  let isDisabled = false;
  const resources = useSelector((state: RootState) => state.integrationsReducer.integrations);
  if (svc === 'Conda') {
    const existingConda = Object.values(resources).find((item) => item.name === 'Conda');
    isDisabled = existingConda !== undefined
  }

  if (integration.category !== category) {
    return null;
  }

  const iconWrapper = (
    <Box
      onClick={() => {
        setShowDialog(integration.activated);
      }}
      sx={{
        width: '64px',
        height: '80px',
        m: 1,
        px: 1,
        py: 1,
        borderRadius: 2,
        //border: `2px solid ${theme.palette.gray['700']}`,
        cursor: integration.activated ? 'pointer' : 'default',
        '&:hover': {
          backgroundColor: integration.activated
            ? theme.palette.gray['300']
            : 'white',
        },
        boxSizing: 'initial',
        backgroundColor: isDisabled ? theme.palette.gray['300'] : theme.palette.gray['25'],
      }}
    >
      <Box
        width="100%"
        maxWidth="100%"
        height="48px"
        display="flex"
        flexDirection="column"
        alignItems="center"
      >
        <IntegrationLogo
          service={service}
          activated={integration.activated}
          size="medium"
        />
      </Box>
      <Typography
        variant={'body1'}
        align={'center'}
        sx={{
          marginTop: '8px',
          color: integration.activated ? 'inherit' : 'grey',
          fontSize: '12px',
        }}
      >
        {service}
      </Typography>
    </Box>
  );


  // For services that require asynchronous connection steps, we show a more realistic message.
  let successMsg = 'Successfully connected to ${service}!'
  if (service === 'Conda' || service === 'Lambda') {
    successMsg = 'Connecting to ${service}...'
  }

  return (
    <Box key={service}>
      <Box>
        {iconWrapper}
        {showDialog && !isDisabled && (
          <IntegrationDialog
            validationSchema={integration.validationSchema}
            dialogContent={integration.dialog}
            user={user}
            service={service}
            onSuccess={() => {
              setShowDialog(false);
              setShowSuccessToast(service);
            }}
            onCloseDialog={() => {
              setShowDialog(false);
              dispatch(resetConnectNewStatus());
            }}
            showMigrationDialog={() => setShowMigrationDialog(true)}
          />
        )}
      </Box>
      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showSuccessToast === service}
        onClose={handleSuccessToastClose}
        key={'integrations-dialog-success-snackbar'}
        autoHideDuration={6000}
      >
        <Alert
          onClose={handleSuccessToastClose}
          severity="success"
          sx={{ width: '100%' }}
        >
          {successMsg}
        </Alert>
      </Snackbar>
    </Box>
  );
};

export default AddIntegrations;
