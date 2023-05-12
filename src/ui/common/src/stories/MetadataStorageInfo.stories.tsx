import { ComponentMeta, ComponentStory } from '@storybook/react';
import React from 'react';

import { ServerConfig } from '../components/pages/account/AccountPage';
import MetadataStorageInfo, {
  FileMetadataStorageInfo,
  GCSMetadataStorageInfo,
  S3MetadataStorageInfo,
} from '../components/pages/account/MetadataStorageInfo';

const mockServerConfig: ServerConfig = {
  aqPath: 'mockAqPath',
  retentionJobPeriod: 'mockRetentionPeriod',
  apiKey: 'mockApiKey',
  storageConfig: {
    type: 'file',
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

const MetadataTemplate: ComponentStory<typeof MetadataStorageInfo> = (args) => (
  <MetadataStorageInfo {...args} />
)

export const MetadataStorageInfoStory = MetadataTemplate.bind({});
MetadataStorageInfoStory.args = {
  serverConfig: mockServerConfig,
};

export default {
  title: 'Components/Metadata Storage Info',
  component: MetadataStorageInfoStory,
  argTypes: {},
} as ComponentMeta<typeof MetadataStorageInfoStory>;

// export const FileMetadataStorageInfoStory: React.FC = () => {
//   const mockFileConfig = {
//     ...mockServerConfig,
//     storageConfig: {
//       type: 'file',
//       ...mockServerConfig.storageConfig,
//     },
//   };

//   return <FileMetadataStorageInfo serverConfig={mockFileConfig} />;
// };

// export const S3MetadataStorageInfoStory: React.FC = () => {
//   const mockFileConfig = {
//     ...mockServerConfig,
//     storageConfig: {
//       type: 's3',
//       ...mockServerConfig.storageConfig,
//     },
//   };

//   return <S3MetadataStorageInfo serverConfig={mockFileConfig} />;
// };

// export const GCSMetadataStorageInfoStory: React.FC = () => {
//   const mockFileConfig = {
//     ...mockServerConfig,
//     storageConfig: {
//       type: 'gcs',
//       ...mockServerConfig.storageConfig,
//     },
//   };

//   return <GCSMetadataStorageInfo serverConfig={mockFileConfig} />;
// };

// const MetadaataStorageInfo