import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import IntegrationLogo from '../../integrations/logo';
import { ServerConfig } from './AccountPage';

interface MetadataPreviewProps {
  serverConfig: ServerConfig;
}
export const FileMetadataStorageInfo: React.FC<MetadataPreviewProps> = ({
  serverConfig,
}) => {
  return (
    <Box sx={{ display: 'flex', height: '85px' }}>
      <Box>
        <IntegrationLogo
          service={'Aqueduct Demo'}
          size={'large'}
          activated={true}
        />
      </Box>
      <Box sx={{ alignSelf: 'center', marginLeft: 2 }}>
        <Typography variant="body1" color={'gray.700'}>
          Storage Type: File
        </Typography>
        <Box sx={{ display: 'flex' }}>
          <Typography variant="body2">
            Location:{' '}
            {serverConfig?.storageConfig?.fileConfig?.directory ||
              'loading ...'}
          </Typography>
        </Box>
      </Box>
    </Box>
  );
};

export const GCSMetadataStorageInfo: React.FC<MetadataPreviewProps> = ({
  serverConfig,
}) => {
  return (
    <Box sx={{ display: 'flex', height: '85px' }}>
      <Box>
        <IntegrationLogo service={'GCS'} size={'large'} activated={true} />
      </Box>
      <Box sx={{ alignSelf: 'center', marginLeft: 2 }}>
        <Typography variant="body1" color={'gray.700'}>
          Storage Type: Google Cloud Storage
        </Typography>
        <Typography variant="body2">
          Bucket:{' '}
          {serverConfig?.storageConfig?.gcsConfig?.bucket || 'loading ...'}
        </Typography>
      </Box>
    </Box>
  );
};

export const S3MetadataStorageInfo: React.FC<MetadataPreviewProps> = ({
  serverConfig,
}) => {
  return (
    <Box sx={{ display: 'flex', height: '85px' }}>
      <Box>
        <IntegrationLogo service={'S3'} size={'large'} activated={true} />
      </Box>
      <Box sx={{ alignSelf: 'center', marginLeft: 2 }}>
        <Typography variant="body1" color={'gray.700'}>
          Storage Type: S3
        </Typography>
        <Box sx={{ display: 'flex' }}>
          <Typography variant="body2">
            Bucket:{' '}
            {serverConfig?.storageConfig?.s3Config?.bucket || 'loading ...'}
          </Typography>
        </Box>
        <Typography variant="body2">
          Region:{' '}
          {serverConfig?.storageConfig?.s3Config?.region || 'loading ...'}
        </Typography>
      </Box>
    </Box>
  );
};

interface MetadataStorageInfoProps {
  serverConfig?: ServerConfig;
}
export const MetadataStorageInfo: React.FC<MetadataStorageInfoProps> = ({
  serverConfig,
}) => {
  // TODO: Show the loading text string here.
  if (!serverConfig) {
    return null;
  }

  let storageInfo;
  switch (serverConfig.storageConfig.type) {
    case 'file': {
      storageInfo = <FileMetadataStorageInfo serverConfig={serverConfig} />;
    }
    case 'gcs': {
      storageInfo = <GCSMetadataStorageInfo serverConfig={serverConfig} />;
    }
    case 's3': {
      storageInfo = <S3MetadataStorageInfo serverConfig={serverConfig} />;
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

export default MetadataStorageInfo;
