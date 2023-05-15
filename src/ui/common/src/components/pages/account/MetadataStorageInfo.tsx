import { CircularProgress } from '@mui/material';
import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import React from 'react';

import { ServerConfig } from '../../../reducers/serverConfig';
import getPathPrefix from '../../../utils/getPathPrefix';
import {
  FilesystemConfig,
  GCSConfig,
  Integration,
  S3Config,
} from '../../../utils/integrations';
import { IntegrationCard } from '../../integrations/cards/card';
import { Card } from '../../layouts/card';

interface MetadataPreviewProps {
  serverConfig: ServerConfig;
}
export const FileMetadataStorageInfo: React.FC<MetadataPreviewProps> = ({
  serverConfig,
}) => {
  if (serverConfig?.storageConfig?.fileConfig === undefined) {
    return <CircularProgress />;
  }

  const filesystemConfig = {
    location: serverConfig?.storageConfig?.fileConfig?.directory,
  };

  const filesystem: Integration = {
    id: '', // This is unused.
    service: 'Filesystem',
    name: serverConfig?.storageConfig?.integration_name,
    config: filesystemConfig as FilesystemConfig,
    createdAt: serverConfig?.storageConfig?.connected_at,
    exec_state: serverConfig?.storageConfig?.exec_state,
  };

  return <IntegrationCard integration={filesystem} numWorkflowsUsingMsg="" />;
};

export const GCSMetadataStorageInfo: React.FC<MetadataPreviewProps> = ({
  serverConfig,
}) => {
  if (serverConfig?.storageConfig?.gcsConfig === undefined) {
    return <CircularProgress />;
  }

  const gcsConfig = {
    bucket: serverConfig?.storageConfig?.gcsConfig?.bucket,
  };

  const gcs: Integration = {
    id: '', // This is unused.
    service: 'GCS',
    name: serverConfig?.storageConfig?.integration_name,
    config: gcsConfig as GCSConfig,

    createdAt: serverConfig?.storageConfig?.connected_at,
    exec_state: serverConfig?.storageConfig?.exec_state,
  };

  return <IntegrationCard integration={gcs} numWorkflowsUsingMsg="" />;
};

export const S3MetadataStorageInfo: React.FC<MetadataPreviewProps> = ({
  serverConfig,
}) => {
  if (serverConfig?.storageConfig?.s3Config === undefined) {
    return <CircularProgress />;
  }

  const s3Config = {
    bucket: serverConfig?.storageConfig?.s3Config?.bucket,
    region: serverConfig?.storageConfig?.s3Config?.region,
  };
  if (serverConfig?.storageConfig?.s3Config.root_dir) {
    s3Config['root_dir'] = serverConfig?.storageConfig?.s3Config.root_dir;
  }

  const s3: Integration = {
    id: '', // This is unused.
    service: 'S3',
    name: serverConfig?.storageConfig?.integration_name,
    config: s3Config as S3Config,

    // This is really "connected at" for storage migration.
    createdAt: serverConfig?.storageConfig?.connected_at,
    exec_state: serverConfig?.storageConfig?.exec_state,
  };

  return <IntegrationCard integration={s3} numWorkflowsUsingMsg="" />;
};

interface MetadataStorageInfoProps {
  serverConfig?: ServerConfig;
}
export const MetadataStorageInfo: React.FC<MetadataStorageInfoProps> = ({
  serverConfig,
}) => {
  if (!serverConfig) {
    return <CircularProgress />;
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
    <Box sx={{ mx: 1, my: 1, display: 'flex', alignItems: 'flex-start' }}>
      <Link
        underline="none"
        color="inherit"
        href={`${getPathPrefix()}/resource/${
          serverConfig?.storageConfig?.integration_id
        }`}
      >
        <Card>{storageInfo}</Card>
      </Link>
    </Box>
  );
};

export default MetadataStorageInfo;
