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
        <Typography
          variant="body1"
          color={'gray.700'}
          fontWeight="fontWeightMedium"
        >
          Storage Type:{' '}
          <Box component="span" fontWeight="fontWeightRegular">
            File
          </Box>
        </Typography>
        <Typography variant="body2" fontWeight="fontWeightMedium">
          Location:{' '}
          <Box component="span" fontWeight="fontWeightRegular">
            {serverConfig?.storageConfig?.fileConfig?.directory ||
              'loading ...'}
          </Box>
        </Typography>
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
        <Typography
          variant="body1"
          color={'gray.700'}
          fontWeight="fontWeightMedium"
        >
          Storage Type:{' '}
          <Box component="span" fontWeight="fontWeightRegular">
            Google Cloud Storage
          </Box>
        </Typography>
        <Typography variant="body2" fontWeight="fontWeightMedium">
          Bucket:{' '}
          <Box component="span" fontWeight="fontWeightRegular">
            {serverConfig?.storageConfig?.gcsConfig?.bucket || 'loading ...'}
          </Box>
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
        <Typography
          variant="body1"
          color={'gray.700'}
          fontWeight="fontWeightBold"
        >
          Storage Type:{' '}
          <Box component="span" fontWeight="fontWeightRegular">
            Amazon S3
          </Box>
        </Typography>
        <Typography variant="body2" fontWeight="fontWeightMedium">
          Bucket:{' '}
          <Box component="span" fontWeight="fontWeightRegular">
            {serverConfig?.storageConfig?.s3Config?.bucket || 'loading ...'}
          </Box>
        </Typography>
        <Typography variant="body2" fontWeight="fontWeightMedium">
          Region:{' '}
          <Box component="span" fontWeight="fontWeightRegular">
            {serverConfig?.storageConfig?.s3Config?.region || 'loading ...'}
          </Box>
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
      break;
    }
    case 'gcs': {
      storageInfo = <GCSMetadataStorageInfo serverConfig={serverConfig} />;
      break;
    }
    case 's3': {
      storageInfo = <S3MetadataStorageInfo serverConfig={serverConfig} />;
      break;
    }
  }

  return (
    <Box>
      <Typography variant="h5" sx={{ mt: 3, mb: 2 }}>
        Metadata Storage
      </Typography>
      {storageInfo}
    </Box>
  );
};

export default MetadataStorageInfo;
