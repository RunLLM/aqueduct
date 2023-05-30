import { Typography } from '@mui/material';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Snackbar from '@mui/material/Snackbar';
import React, { useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { resetConnectNewStatus } from '../../reducers/resource';
import { AppDispatch, RootState } from '../../stores/store';
import { theme } from '../../styles/theme/theme';
import UserProfile from '../../utils/auth';
import { Info, Service, ServiceInfoMap } from '../../utils/resources';
import ResourceDialog from './dialogs/dialog';
import ResourceLogo from './logo';

type Props = {
  user: UserProfile;
  supportedResources: ServiceInfoMap;
  category: string;
};

const AddResources: React.FC<Props> = ({
  user,
  supportedResources,
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
        {Object.entries(supportedResources)
          .filter(([svc]) => svc !== 'Aqueduct')
          .sort(([name1], [name2]) => name1.localeCompare(name2))
          .map(([svc, resource]) => {
            return (
              <AddResourceListItem
                key={svc as string}
                svc={svc}
                dialog={resource.dialog}
                resource={resource}
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

interface AddResourceListItemProps {
  svc: string;
  resource: Info;
  category: string;
  user: UserProfile;
  showSuccessToast: string;
  // callback functions
  handleSuccessToastClose: () => void;
  setShowSuccessToast: React.Dispatch<React.SetStateAction<Service>>;
  setShowMigrationDialog: React.Dispatch<React.SetStateAction<boolean>>;
  dialog: React.FC;
}

const AddResourceListItem: React.FC<AddResourceListItemProps> = ({
  svc,
  resource,
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

  // If this is a conda resource, check if it has already been registered.
  // If it has, disable the new Conda button.
  const resources = useSelector(
    (state: RootState) => state.resourcesReducer.resources
  );
  if (svc === 'Conda') {
    const existingConda = Object.values(resources).find(
      (item) => item.name === 'Conda'
    );
    resource.activated = existingConda === undefined;
  }

  if (resource.category !== category) {
    return null;
  }

  const iconWrapper = (
    <Box
      onClick={() => {
        setShowDialog(resource.activated);
      }}
      sx={{
        width: '64px',
        height: '80px',
        mr: 1,
        my: 1,
        px: 1,
        py: 1,
        borderRadius: 2,
        //border: `2px solid ${theme.palette.gray['700']}`,
        cursor: resource.activated ? 'pointer' : 'default',
        '&:hover': {
          backgroundColor: resource.activated
            ? theme.palette.gray['300']
            : 'white',
        },
        boxSizing: 'initial',
        backgroundColor: theme.palette.gray['25'],
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
        <ResourceLogo
          service={service}
          activated={resource.activated}
          size="medium"
        />
      </Box>
      <Typography
        variant={'body1'}
        align={'center'}
        sx={{
          marginTop: '8px',
          color: resource.activated ? 'inherit' : 'grey',
          fontSize: '12px',
        }}
      >
        {service}
      </Typography>
    </Box>
  );

  // For services that require asynchronous connection steps, we show a more realistic message.
  let successMsg = `Successfully connected to ${service}!`;
  if (service === 'Conda' || service === 'Lambda') {
    successMsg = `Connecting to ${service}...`;
  }

  return (
    <Box key={service}>
      <Box>
        {iconWrapper}
        {showDialog && (
          <ResourceDialog
            validationSchema={resource.validationSchema}
            dialogContent={resource.dialog}
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
        key={'resources-dialog-success-snackbar'}
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

export default AddResources;
