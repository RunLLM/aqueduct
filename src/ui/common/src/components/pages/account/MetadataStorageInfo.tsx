import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

import { ServerConfig } from '../../../reducers/serverConfig';
import IntegrationLogo from '../../integrations/logo';
import {FilesystemConfig, Integration, S3Config} from "../../../utils/integrations";
import {CircularProgress} from "@mui/material";
import {IntegrationCard} from "../../integrations/cards/card";
import {Card} from "../../layouts/card";
import Link from "@mui/material/Link";
import getPathPrefix from "../../../utils/getPathPrefix";

interface MetadataPreviewProps {
  serverConfig: ServerConfig;
}
export const FileMetadataStorageInfo: React.FC<MetadataPreviewProps> = ({
  serverConfig,
}) => {
  if (serverConfig?.storageConfig?.fileConfig === undefined) {
    return <CircularProgress/>
  }

  const filesystemConfig = {
    location: serverConfig?.storageConfig?.fileConfig?.directory,
  }

  const filesystem: Integration = {
    id: '', // This is unused.
    service: 'Filesystem',
    name: serverConfig?.storageConfig?.integration_name,
    config: filesystemConfig as FilesystemConfig,
    createdAt: 0,
    exec_state: serverConfig?.storageConfig?.exec_state,
  }

  return <IntegrationCard integration={filesystem} numWorkflowsUsingMsg='' />
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
          fontWeight="fontWeightBold"
        >
          Storage Type:{' '}
          <Box component="span" fontWeight="fontWeightRegular">
            Google Cloud Storage
          </Box>
        </Typography>

        <Typography variant="body2" fontWeight="fontWeightRegular">
          Name:{' '}
          <Box component="span" fontWeight="fontWeightRegular">
            {serverConfig?.storageConfig?.integration_name || 'loading ...'}
          </Box>
        </Typography>
        <Typography variant="body2" fontWeight="fontWeightRegular">
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
  if (serverConfig?.storageConfig?.s3Config === undefined) {
    return <CircularProgress/>
  }

  const s3Config = {
    bucket: serverConfig?.storageConfig?.s3Config?.bucket,
    region: serverConfig?.storageConfig?.s3Config?.region,
  }
  if (serverConfig?.storageConfig?.s3Config.root_dir) {
    s3Config['root_dir'] = serverConfig?.storageConfig?.s3Config.root_dir
  }

  const s3: Integration = {
    id: '', // This is unused.
    service: 'S3',
    name: serverConfig?.storageConfig?.integration_name,
    config: s3Config as S3Config,

    // This is really "connected at" for storage migration.
    createdAt: serverConfig?.storageConfig?.connected_at || 0,
    exec_state: serverConfig?.storageConfig?.exec_state,
  }

  return <IntegrationCard integration={s3} numWorkflowsUsingMsg='' />
};

interface MetadataStorageInfoProps {
  serverConfig?: ServerConfig;
}
export const MetadataStorageInfo: React.FC<MetadataStorageInfoProps> = ({
  serverConfig,
}) => {
  if (!serverConfig) {
    return <CircularProgress/>;
  }

  let storageInfo;
  let detailsLink: string | undefined = undefined;
  switch (serverConfig.storageConfig.type) {
    case 'file': {
      storageInfo = <FileMetadataStorageInfo serverConfig={serverConfig} />;
      break;
    }
    case 'gcs': {
      storageInfo = <GCSMetadataStorageInfo serverConfig={serverConfig} />;
      detailsLink = `${getPathPrefix()}/resource/${serverConfig?.storageConfig?.integration_id}`
      break;
    }
    case 's3': {
      storageInfo = <S3MetadataStorageInfo serverConfig={serverConfig} />;
      detailsLink = `${getPathPrefix()}/resource/${serverConfig?.storageConfig?.integration_id}`
      break;
    }
  }

  return (
      <Box sx={{ mx: 1, my: 1 }}>
      <Link
          underline="none"
          color="inherit"
          href={detailsLink}
      >
        <Card>
          {storageInfo}
        </Card>
      </Link>
    </Box>
  )
};

export default MetadataStorageInfo;
