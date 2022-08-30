from enum import Enum
from typing import Optional

from aqueduct_executor.operators.utils.enums import MetaEnum
from pydantic import BaseModel


class StorageType(str, Enum, metaclass=MetaEnum):
    S3 = "s3"
    File = "file"
    GCS = "gcs"


class FileStorageConfig(BaseModel):
    directory: str


class S3StorageConfig(BaseModel):
    region: str
    bucket: str
    credentials_path: str
    credentials_profile: str
    aws_access_key_id: str = ""
    aws_secret_access_key: str = ""


class GCSStorageConfig(BaseModel):
    bucket: str
    service_account_credentials: str


class StorageConfig(BaseModel):
    type: StorageType
    file_config: Optional[FileStorageConfig] = None
    s3_config: Optional[S3StorageConfig] = None
    gcs_config: Optional[GCSStorageConfig] = None
