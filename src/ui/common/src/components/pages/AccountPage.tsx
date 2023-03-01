import { Alert, CircularProgress, Snackbar } from '@mui/material';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { handleLoadIntegrations } from '../../reducers/integrations';
import { AppDispatch, RootState } from '../../stores/store';
import UserProfile from '../../utils/auth';
import {
  Integration,
  IntegrationCategories,
  SupportedIntegrations,
} from '../../utils/integrations';
import {
  isFailed,
  isInitial,
  isLoading,
  isSucceeded,
} from '../../utils/shared';
import { CodeBlock } from '../CodeBlock';
import { useAqueductConsts } from '../hooks/useAqueductConsts';
import IntegrationLogo from '../integrations/logo';
import DefaultLayout from '../layouts/default';
import { BreadcrumbLink } from '../layouts/NavBar';
import AccountNotificationSettingsSelector, {
  NotificationConfigsMap,
} from '../notifications/AccountNotificationSettingsSelector';
import { LayoutProps } from './types';

type AccountPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

type ServerConfig = {
  aqPath: string;
  retentionJobPeriod: string;
  apiKey: string;
  storageConfig: {
    type: string;
    fileConfig?: {
      directory: string;
    };
    gcsConfig?: {
      bucket: string;
    };
    s3Config?: {
      region: string;
      bucket: string;
    };
  };
};

async function getServerConfig(
  apiAddress: string,
  apiKey: string
): Promise<ServerConfig> {
  try {
    const configRequest = await fetch(`${apiAddress}/api/config`, {
      method: 'GET',
      headers: {
        'api-key': apiKey,
      },
    });

    const responseBody = await configRequest.json();

    if (!configRequest.ok) {
      console.log('Error fetching config');
    }

    return responseBody as ServerConfig;
  } catch (error) {
    console.log('config fetch error: ', error);
  }
}

// `UpdateNotifications` attempts to update all notification integration by calling
// `integration/<id>/edit` route separately. It returns an error message if any error occurs.
// Otherwise, the message will be empty.
async function UpdateNotifications(
  apiAddress: string,
  apiKey: string,
  integrations: { [id: string]: Integration },
  configs: NotificationConfigsMap
): Promise<string> {
  const promiseResults = Object.entries(configs).map(async ([id, config]) => {
    try {
      const res = await fetch(`${apiAddress}/api/integration/${id}/edit`, {
        method: 'POST',
        headers: {
          'api-key': apiKey,
          'integration-name': integrations[id]?.name ?? '',
          'integration-config': JSON.stringify(config),
        },
      });

      const responseBody = await res.json();

      if (!res.ok) {
        const msg = responseBody.error as string;
        return `Failed to update ${id}: ${msg} .`;
      }

      return '';
    } catch (error) {
      const msg = error as string;
      return `Failed to update ${id}: ${msg} .`;
    }
  });

  const results = await Promise.all(promiseResults);
  // combine error messages
  return results.filter((x) => !!x).join('\n');
}

interface MetadataStorageInfoProps {
  serverConfig?: ServerConfig;
}

const MetadataStorageInfo: React.FC<MetadataStorageInfoProps> = ({
  serverConfig,
}) => {
  // TODO: Show the loading text string here.
  if (!serverConfig) {
    return null;
  }

  let storageInfo;
  // 85px here is the logo height.
  const fileMetadataStorageInfo = (
    <Box sx={{ display: 'flex', height: '85px' }}>
      <Box>
        <IntegrationLogo
          service={'Aqueduct Demo'}
          size={'large'}
          activated={true}
        />
      </Box>
      <Box sx={{ alignSelf: 'center', marginLeft: 2 }}>
        <Typography variant="body2" color={'gray.700'}>
          Storage Type: {serverConfig?.storageConfig?.type || 'loading ...'}
        </Typography>
        <Typography variant="body1">
          Location:{' '}
          {serverConfig?.storageConfig?.fileConfig?.directory || 'loading ...'}
        </Typography>
      </Box>
    </Box>
  );

  const gcsMetadataStorageInfo = (
    <Box sx={{ display: 'flex', height: '85px' }}>
      <Box>
        <IntegrationLogo service={'GCS'} size={'large'} activated={true} />
      </Box>
      <Box sx={{ alignSelf: 'center', marginLeft: 2 }}>
        <Typography variant="body2" color={'gray.700'}>
          Storage Type: {serverConfig?.storageConfig?.type || 'loading ...'}
        </Typography>
        <Typography variant="body1">
          Bucket:{' '}
          {serverConfig?.storageConfig?.gcsConfig?.bucket || 'loading ...'}
        </Typography>
      </Box>
    </Box>
  );

  const s3MetadataStorageInfo = (
    <Box sx={{ display: 'flex', height: '85px' }}>
      <Box>
        <IntegrationLogo service={'S3'} size={'large'} activated={true} />
      </Box>
      <Box sx={{ alignSelf: 'center', marginLeft: 2 }}>
        <Typography variant="body2" color={'gray.700'}>
          Storage Type: {serverConfig?.storageConfig?.type || 'loading ...'}
        </Typography>
        <Typography variant="body1">
          Bucket:{' '}
          {serverConfig?.storageConfig?.s3Config?.bucket || 'loading ...'}
        </Typography>
        <Typography variant="body1">
          Region:{' '}
          {serverConfig?.storageConfig?.s3Config?.region || 'loading ...'}
        </Typography>
      </Box>
    </Box>
  );

  switch (serverConfig.storageConfig.type) {
    case 'file': {
      storageInfo = fileMetadataStorageInfo;
      break;
    }
    case 'gcs': {
      storageInfo = gcsMetadataStorageInfo;
      break;
    }
    case 's3': {
      storageInfo = s3MetadataStorageInfo;
      break;
    }
  }

  return (
    <Box>
      <Typography variant="h5" sx={{ mt: 3 }}>
        Metadata Storage
      </Typography>
      {storageInfo}
    </Box>
  );
};

const AccountPage: React.FC<AccountPageProps> = ({
  user,
  Layout = DefaultLayout,
}) => {
  // Set the title of the page on page load.
  useEffect(() => {
    document.title = 'Account | Aqueduct';
  }, []);

  const { apiAddress } = useAqueductConsts();
  const serverAddress = apiAddress ? `${apiAddress}` : '<server address>';
  const apiConnectionSnippet = `import aqueduct
client = aqueduct.Client(
    "${user.apiKey}",
    "${serverAddress}"
)`;
  const dispatch: AppDispatch = useDispatch();
  const maxContentWidth = '600px';
  const integrationsReducer = useSelector(
    (state: RootState) => state.integrationsReducer
  );

  const [serverConfig, setServerConfig] = useState<ServerConfig | null>(null);
  const notifications = Object.values(integrationsReducer.integrations).filter(
    (x) =>
      SupportedIntegrations[x.service].category ===
      IntegrationCategories.NOTIFICATION
  );

  const [updatingNotifications, setUpdatingNotifications] = useState(false);
  const [notificationUpdateError, setNotificationUpdateError] = useState('');
  const [showNotificationUpdateSnackbar, setShowNotificationUpdateSnackbar] =
    useState(false);

  useEffect(() => {
    async function fetchServerConfig() {
      const serverConfig = await getServerConfig(apiAddress, user.apiKey);
      setServerConfig(serverConfig);
    }

    fetchServerConfig();
  }, []);

  useEffect(() => {
    if (!updatingNotifications) {
      dispatch(
        handleLoadIntegrations({ apiKey: user.apiKey, forceLoad: true })
      );
    }
  }, [updatingNotifications, dispatch, user.apiKey]);

  let notificationSection = null;
  if (
    isLoading(integrationsReducer.status) ||
    isInitial(integrationsReducer.status)
  ) {
    notificationSection = <CircularProgress />;
  }

  if (isFailed(integrationsReducer.status)) {
    notificationSection = (
      <Alert title={integrationsReducer.status.err} severity="error" />
    );
  }

  if (isSucceeded(integrationsReducer.status)) {
    notificationSection = (
      <AccountNotificationSettingsSelector
        notifications={notifications}
        onSave={async (configs) => {
          setUpdatingNotifications(true);
          const err = await UpdateNotifications(
            apiAddress,
            user.apiKey,
            integrationsReducer.integrations,
            configs
          );
          setNotificationUpdateError(err);
          setUpdatingNotifications(false);
          setShowNotificationUpdateSnackbar(true);
        }}
        isSaving={updatingNotifications}
      />
    );
  }
  return (
    <Layout
      breadcrumbs={[BreadcrumbLink.HOME, BreadcrumbLink.ACCOUNT]}
      user={user}
    >
      <Typography variant="h5">API Key</Typography>
      <Box sx={{ my: 1 }}>
        <code>{user.apiKey}</code>
      </Box>

      <Box
        sx={{
          display: 'flex',
          flexDirection: 'column',
          width: maxContentWidth,
        }}
      >
        <Typography variant="body1" sx={{ fontWeight: 'bold', mr: '8px' }}>
          Python SDK Connection Snippet
        </Typography>
        <Box
          sx={{
            marginTop: '8px',
          }}
        >
          <CodeBlock language="python">{apiConnectionSnippet}</CodeBlock>
        </Box>
      </Box>
      <Typography variant="h5" sx={{ mt: 3 }}>
        Notifications
      </Typography>
      {notifications.length > 0 && (
        <Typography variant="body2" marginBottom={1}>
          Configure how your notification(s) apply to all workflows. You can
          override these settings in for individual workflows in workflow
          settings page.
        </Typography>
      )}
      {notificationSection}

      <MetadataStorageInfo serverConfig={serverConfig} />

      <Snackbar
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
        open={showNotificationUpdateSnackbar}
        key="notification-update-err-snackbar"
        autoHideDuration={6000}
        onClose={() => {
          setShowNotificationUpdateSnackbar(false);
        }}
      >
        <Alert
          severity={!notificationUpdateError ? 'success' : 'error'}
          sx={{ width: '100%' }}
        >
          {!notificationUpdateError
            ? 'Successfully updated notification settings.'
            : notificationUpdateError}
        </Alert>
      </Snackbar>
    </Layout>
  );
};

export default AccountPage;