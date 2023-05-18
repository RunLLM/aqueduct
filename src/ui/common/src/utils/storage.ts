import { IntegrationConfig, Service } from './resources';

export enum StorageType {
  S3 = 's3',
  File = 'file',
  GCS = 'gcs',
}

export const StorageTypeNames = {
  s3: 'AWS S3',
  file: 'Local File System',
  gcs: 'Google Cloud Storage',
};

export type S3Config = {
  region: string;
  bucket: string;
  credentials_path?: string;
  credentials_profile?: string;
  aws_access_key_id?: string;
  aws_secret_access_key?: string;
};

export type FileConfig = {
  directory: string;
};

export type GCSConfig = {
  bucket: string;
  service_account_credentials?: string;
};

export type StorageConfig = {
  type: StorageType;
  s3_config?: S3Config;
  file_config?: FileConfig;
  gcs_config?: GCSConfig;
};

export type MetadataStorageConfig = {
  type: StorageType;
  s3Config?: S3Config;
  fileConfig?: FileConfig;
  gcsConfig?: GCSConfig;
};

export type ServerConfig = {
  aqPath: string;
  retentionJobPeriod: string;
  apiKey: string;
  storageConfig: MetadataStorageConfig;
};

function convertS3IntegrationtoMetadataStorageConfig(
  storage: S3Config
): MetadataStorageConfig {
  return {
    type: StorageType.S3,
    s3Config: {
      region: storage.region,
      bucket: 's3://' + storage.bucket,
    },
  };
}

function convertGCSIntegrationtoMetadataStorageConfig(
  storage: GCSConfig
): MetadataStorageConfig {
  return {
    type: StorageType.GCS,
    gcsConfig: {
      bucket: storage.bucket,
    },
  };
}

export function convertIntegrationConfigToServerConfig(
  storage: IntegrationConfig,
  metadataStorage: ServerConfig,
  service: Service
): ServerConfig {
  let storageConfig;
  switch (service) {
    case 'S3':
      storageConfig = convertS3IntegrationtoMetadataStorageConfig(
        storage as S3Config
      );
      break;
    case 'GCS':
      storageConfig = convertGCSIntegrationtoMetadataStorageConfig(
        storage as GCSConfig
      );
      break;
    default:
      return null;
  }
  return {
    aqPath: metadataStorage.aqPath,
    retentionJobPeriod: metadataStorage.retentionJobPeriod,
    apiKey: metadataStorage.apiKey,
    storageConfig: storageConfig,
  };
}
