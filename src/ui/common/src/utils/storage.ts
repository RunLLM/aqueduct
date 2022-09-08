export enum StorageType {
  S3 = 's3',
  File = 'file',
  GCS = 'gcs',
}

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
