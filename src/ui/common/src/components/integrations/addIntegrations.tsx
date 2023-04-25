import { Typography } from '@mui/material';
import { DialogActions, DialogContent } from '@mui/material';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import Snackbar from '@mui/material/Snackbar';
import React, { useState } from 'react';
import { useDispatch } from 'react-redux';

import { resetConnectNewStatus } from '../../reducers/integration';
import { AppDispatch } from '../../stores/store';
import { theme } from '../../styles/theme/theme';
import UserProfile from '../../utils/auth';
import {
  Info,
  Service,
  ServiceInfoMap,
  SupportedIntegrations,
} from '../../utils/integrations';
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

  // TODO: Add dialog component to the integrationobject that's part of the supportedIntegrations array.
  // This will let us easily choose which component to use when rendering the dialog.
  // Not much "inheritence" to be used here, but we can have a common interface for IntegrationDialogs.
  console.log('addIntegrations supportedIntegrations: ', supportedIntegrations);

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

  const [showKubernetesDialog, setShowKubernetesDialog] = useState(false);
  const [showOndemandDialog, setShowOndemandDialog] = useState(false);
  const [showSelectProviderDialog, setShowSelectProviderDialog] =
    useState(false);

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
        backgroundColor: '#F8F8F8', // gray/light2
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

  if (service !== 'Kubernetes') {
    return (
      <Box key={service}>
        <Box>
          {iconWrapper}
          {showDialog && (
            <IntegrationDialog
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
            {`Successfully connected to ${service}!`}
          </Alert>
        </Snackbar>
      </Box>
    );
  }

  const handleRegularK8s = () => {
    setShowKubernetesDialog(true);
    setShowDialog(false);
  };

  const handleOndemandK8s = () => {
    setShowSelectProviderDialog(true);
    setShowDialog(false);
  };

  const handlePrevious = () => {
    setShowSelectProviderDialog(false);
    setShowDialog(true);
  };

  const handleAWSClick = () => {
    setShowOndemandDialog(true);
    setShowSelectProviderDialog(false);
  };

  return (
    <Box key={service}>
      <Box>
        {iconWrapper}
        <Dialog open={showDialog} onClose={() => setShowDialog(false)}>
          <DialogTitle
            sx={{ display: 'flex', alignItems: 'center', gap: '8px' }}
          >
            <div>
              <Typography variant="h5" sx={{ color: 'black' }}>
                Connect to Kubernetes
              </Typography>
            </div>
          </DialogTitle>
          <DialogContent sx={{ marginTop: '8px' }}>
            <Button
              sx={{
                textTransform: 'none',
                marginBottom: '8px',
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
              }}
              onClick={handleRegularK8s}
            >
              <IntegrationLogo
                service={service}
                activated={integration.activated}
                size="small"
              />
              <Typography
                variant="body2"
                sx={{ color: 'black', fontSize: '18px' }}
              >
                I have an existing Kubernetes cluster I&apos;d like to use
              </Typography>
            </Button>
            <Button
              sx={{
                textTransform: 'none',
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
              }}
              onClick={handleOndemandK8s}
            >
              <IntegrationLogo
                service={'Aqueduct'}
                activated={SupportedIntegrations['Aqueduct'].activated}
                size="small"
              />
              <Typography
                variant="body2"
                sx={{ color: 'black', fontSize: '18px' }}
              >
                I&apos;d like Aqueduct to create & manage a cluster for me
              </Typography>
            </Button>
          </DialogContent>
          <DialogActions>
            <Button autoFocus onClick={() => setShowDialog(false)}>
              Cancel
            </Button>
          </DialogActions>
        </Dialog>

        {showKubernetesDialog && (
          <IntegrationDialog
            validationSchema={integration.validationSchema}
            user={user}
            dialogContent={dialog}
            service={service}
            onSuccess={() => {
              setShowKubernetesDialog(false);
              setShowSuccessToast(service);
            }}
            onCloseDialog={() => {
              setShowKubernetesDialog(false);
              dispatch(resetConnectNewStatus());
            }}
            showMigrationDialog={() => setShowMigrationDialog(true)}
          />
        )}

        <Dialog
          open={showSelectProviderDialog}
          onClose={() => setShowSelectProviderDialog(false)}
        >
          <DialogTitle
            sx={{ display: 'flex', alignItems: 'center', gap: '8px' }}
          >
            <IntegrationLogo
              service={'Aqueduct'}
              activated={SupportedIntegrations['Aqueduct'].activated}
              size="small"
            />
            <div>
              <Typography variant="h5" sx={{ color: 'black' }}>
                +
              </Typography>
            </div>
            <IntegrationLogo
              service={service}
              activated={integration.activated}
              size="small"
            />
            <div>
              <Typography variant="h5" sx={{ color: 'black' }}>
                Aqueduct-managed Kubernetes
              </Typography>
            </div>
          </DialogTitle>
          <DialogContent
            sx={{
              display: 'flex',
              alignItems: 'center',
              paddingLeft: '54px',
              gap: '32px',
              marginTop: '16px',
              '& button': { backgroundColor: '#F8F8F8' },
            }}
          >
            <Button onClick={handleAWSClick}>
              <IntegrationLogo
                service={'Amazon'}
                activated={SupportedIntegrations['Amazon'].activated}
                size="large"
              />
            </Button>
            <Button disabled={true}>
              <IntegrationLogo
                service={'GCP'}
                activated={SupportedIntegrations['GCP'].activated}
                size="large"
              />
            </Button>
            <Button disabled={true}>
              <IntegrationLogo
                service={'Azure'}
                activated={SupportedIntegrations['Azure'].activated}
                size="large"
              />
            </Button>
          </DialogContent>
          <DialogActions>
            <Button autoFocus onClick={handlePrevious}>
              Previous
            </Button>
            <Button
              autoFocus
              onClick={() => setShowSelectProviderDialog(false)}
            >
              Cancel
            </Button>
          </DialogActions>
        </Dialog>

        {showOndemandDialog && (
          <IntegrationDialog
            user={user}
            service="AWS"
            onSuccess={() => {
              setShowOndemandDialog(false);
              setShowSuccessToast(service);
            }}
            onCloseDialog={() => {
              setShowOndemandDialog(false);
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
          {`Successfully connected to ${service}!`}
        </Alert>
      </Snackbar>
    </Box>
  );
};

export default AddIntegrations;
