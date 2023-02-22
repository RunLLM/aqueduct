import React from 'react';

import { ServerConfig } from '../components/pages/account/AccountPage';
import MetadataStorageInfo, {
  FileMetadataStorageInfo,
} from '../components/pages/account/MetadataStorageInfo';

const mockServerConfig: ServerConfig = {
  aqPath: 'mockAqPath',
  retentionJobPeriod: 'mockRetentionPeriod',
  apiKey: 'mockApiKey',
  storageConfig: {
    type: 'gcs',
    fileConfig: {
      directory: '/storybook/metadataStorageInfoStory.tsx',
    },
    s3Config: {
      bucket: 's3-mock-storybook-bucket',
      region: 'us-east-2',
    },
    gcsConfig: {
      bucket: 'gcs-mock-storybook-bucket',
    },
  },
};
export const MetadataStorageInfoStory: React.FC = () => {
  return <MetadataStorageInfo serverConfig={mockServerConfig} />;
};

export const FileMetadataStorageInfoStory: React.FC = () => {
  const mockFileConfig = {
    ...mockServerConfig,
    storageConfig: {
      type: 'file',
      ...mockServerConfig.storageConfig,
    },
  };

  console.log('mockFileConfig: ', mockFileConfig);

  return <FileMetadataStorageInfo serverConfig={mockFileConfig} />;
};

export default MetadataStorageInfoStory;
