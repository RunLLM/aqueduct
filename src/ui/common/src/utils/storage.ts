import { IntegrationConfig, Service } from './integrations';

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
  credentials_path: string;
  credentials_profile: string;
  aws_access_key_id: string;
  aws_secret_access_key: string;
};

export type FileConfig = {
  directory: string;
};

export type GCSConfig = {
  bucket: string;
  service_account_credentials: string;
};

export type StorageConfig = {
  type: StorageType;
  s3_config?: S3Config;
  file_config?: FileConfig;
  gcs_config?: GCSConfig;
};

export type MetadataStorageConfig = {
  aqPath: string;
  retentionJobPeriod: string;
  apiKey: string;
  storageConfig: {
    type: StorageType;
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

export async function getMetadataStorageConfig(
  apiAddress: string,
  apiKey: string
): Promise<MetadataStorageConfig> {
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

    return responseBody as MetadataStorageConfig;
  } catch (error) {
    console.log('config fetch error: ', error);
  }
}

function convertS3IntegrationtoStorageConfig(
  storage: S3Config,
  metadataStorage: MetadataStorageConfig
): MetadataStorageConfig {
  return {
    aqPath: metadataStorage.aqPath,
    retentionJobPeriod: metadataStorage.retentionJobPeriod,
    apiKey: metadataStorage.apiKey,
    storageConfig: {
      type: StorageType.S3,
      s3Config: {
        region: storage.region,
        bucket: 's3://' + storage.bucket,
      },
    },
  };
}

function convertGCSIntegrationtoStorageConfig(
  storage: GCSConfig,
  metadataStorage: MetadataStorageConfig
): MetadataStorageConfig {
  return {
    aqPath: metadataStorage.aqPath,
    retentionJobPeriod: metadataStorage.retentionJobPeriod,
    apiKey: metadataStorage.apiKey,
    storageConfig: {
      type: StorageType.GCS,
      gcsConfig: {
        bucket: storage.bucket,
      },
    },
  };
}

export function convertIntegrationConfigToMetadataStorageConfig(
  storage: IntegrationConfig,
  metadataStorage: MetadataStorageConfig,
  service: Service
): MetadataStorageConfig {
  switch (service) {
    case 'S3':
      return convertS3IntegrationtoStorageConfig(
        storage as S3Config,
        metadataStorage
      );
    case 'GCS':
      return convertGCSIntegrationtoStorageConfig(
        storage as GCSConfig,
        metadataStorage
      );
    default:
      return null;
  }
}
