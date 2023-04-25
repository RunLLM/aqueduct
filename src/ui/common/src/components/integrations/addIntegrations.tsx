import { Typography } from '@mui/material';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Snackbar from '@mui/material/Snackbar';
import React, { useState } from 'react';
import { useDispatch } from 'react-redux';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import {
  DialogActions,
  DialogContent,
} from '@mui/material';

import { resetConnectNewStatus } from '../../reducers/integration';
import { AppDispatch } from '../../stores/store';
import { theme } from '../../styles/theme/theme';
import UserProfile from '../../utils/auth';
import { Info, Service, ServiceInfoMap } from '../../utils/integrations';
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
          .filter(([svc]) => svc !== 'Aqueduct Demo')
          .sort(([name1], [name2]) => name1.localeCompare(name2))
          .map(([svc, integration]) => {
            return (
              <AddIntegrationListItem
                key={svc as string}
                svc={svc}
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
}) => {
  const dispatch: AppDispatch = useDispatch();
  const service = svc as Service;
  const [showDialog, setShowDialog] = useState(false);

  if (integration.category !== category) {
    return null;
  }

  const iconWrapper = (
    <Box
      onClick={() => {
        setShowDialog(integration.activated);
      }}
      sx={{
        width: '160px',
        height: '128px',
        m: 1,
        px: 2,
        py: 2,
        borderRadius: 2,
        border: `2px solid ${theme.palette.gray['700']}`,
        cursor: integration.activated ? 'pointer' : 'default',
        '&:hover': {
          backgroundColor: integration.activated
            ? theme.palette.gray['300']
            : 'white',
        },
      }}
    >
      <Box
        width="160px"
        maxWidth="160px"
        display="flex"
        flexDirection="column"
        alignItems="center"
      >
        <IntegrationLogo
          service={service}
          activated={integration.activated}
          size="large"
        />
      </Box>
      <Typography
        variant={'body1'}
        align={'center'}
        sx={{
          marginTop: '16px',
          color: integration.activated ? 'inherit' : 'grey',
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
  } else {
    const [showKubernetesDialog, setShowKubernetesDialog] = useState(false);
    const [showOndemandDialog, setShowOndemandDialog] = useState(false);
    const [showSelectProviderDialog, setShowSelectProviderDialog] = useState(false);
  
    const handleOption1Click = () => {
      setShowKubernetesDialog(true);
      setShowDialog(false);
    };
  
    const handleOption2Click = () => {
      setShowSelectProviderDialog(true);
      setShowDialog(false);
    };

    const handleAwsClick = () => {
      setShowOndemandDialog(true);
      setShowSelectProviderDialog(false);
    };

    return (
      <Box key={service}>
        <Box>
          {iconWrapper}
          <>
            <Dialog open={showDialog} onClose={() => setShowDialog(false)}>
              <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <IntegrationLogo
                  service={service}
                  activated={integration.activated}
                  size="small"
                />
                <Typography variant="h5" sx={{ color: 'black' }}>
                  Connect to Kubernetes
                </Typography>
              </DialogTitle>
              <DialogContent>
                <Button sx={{textTransform: 'none', marginBottom: '12px', display: 'flex', gap: '8px'}} onClick={handleOption1Click}>
                  <IntegrationLogo
                    service={service}
                    activated={integration.activated}
                    size="small"
                  />
                  <Typography variant="body2" sx={{ color: 'black', fontSize: '20px' }}>
                    I have an existing Kubernetes cluster I'd like to use
                  </Typography>
                </Button>
                <Button sx={{textTransform: 'none', display: 'flex', gap: '8px'}} onClick={handleOption2Click}>
                  <IntegrationLogo
                      service={'Aqueduct Demo'}
                      activated={integration.activated}
                      size="small"
                  />
                  <Typography variant="body2" sx={{ color: 'black', fontSize: '20px' }}>
                    I'd like Aqueduct to create & manage a cluster for me
                  </Typography>
                </Button>
              </DialogContent>
            </Dialog>

            {showKubernetesDialog && (
              <IntegrationDialog
                user={user}
                service={service}
                onSuccess={() => {
                  setShowKubernetesDialog(false)
                  setShowSuccessToast(service);
                }}
                onCloseDialog={() => {
                  setShowKubernetesDialog(false)
                  dispatch(resetConnectNewStatus());
                }}
                showMigrationDialog={() => setShowMigrationDialog(true)}
              />
            )}

            <Dialog open={showSelectProviderDialog} onClose={() => setShowSelectProviderDialog(false)}>
              <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <IntegrationLogo
                      service={service}
                      activated={integration.activated}
                      size="small"
                />
                <Typography variant="h5" sx={{ color: 'black' }}>
                  +
                </Typography>
                <IntegrationLogo
                      service={'Aqueduct Demo'}
                      activated={integration.activated}
                      size="small"
                />
                <Typography variant="h5" sx={{ color: 'black' }}>
                Aqueduct-managed Kubernetes
                </Typography>
              </DialogTitle>
              <DialogContent sx={{ display: 'flex', alignItems: 'center', paddingLeft: '54px', gap: '32px', '& button': { backgroundColor: '#F8F8F8' } }}>
                <Button onClick={handleAwsClick}>
                  <IntegrationLogo
                        service={'AWS'}
                        activated={integration.activated}
                        size="large"
                  />
                </Button>
                <Button disabled={true}>
                  <IntegrationLogo
                        service={'GCP'}
                        activated={false}
                        size="large"
                  />
                </Button>
                <Button disabled={true}>
                  <IntegrationLogo
                        service={'Azure'}
                        activated={false}
                        size="large"
                  />
                </Button>
              </DialogContent>
            </Dialog>

            {showOndemandDialog && (
            <IntegrationDialog
              user={user}
              service='AWS'
              onSuccess={() => {
                setShowOndemandDialog(false)
                setShowSuccessToast(service);
              }}
              onCloseDialog={() => {
                setShowOndemandDialog(false)
                dispatch(resetConnectNewStatus());
              }}
              showMigrationDialog={() => setShowMigrationDialog(true)}
            />
            )}
          </>
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
};

export default AddIntegrations;
