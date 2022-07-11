from enum import Enum
from typing import Optional

from aqueduct_executor.operators.utils.enums import MetaEnum
from pydantic import BaseModel


class StorageType(str, Enum, metaclass=MetaEnum):
    S3 = "s3"
    File = "file"


class FileStorageConfig(BaseModel):
    directory: str


class S3StorageConfig(BaseModel):
    region: str
    bucket: str
    credentials_path: str
    credentials_profile: str


class StorageConfig(BaseModel):
    type: StorageType
    file_config: Optional[FileStorageConfig] = None
    s3_config: Optional[S3StorageConfig] = None
