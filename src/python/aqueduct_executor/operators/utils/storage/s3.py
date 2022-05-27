from typing import Any

import boto3
from botocore.config import Config as BotoConfig

from typing import Any

from aqueduct_executor.operators.utils.storage.config import S3StorageConfig
from aqueduct_executor.operators.utils.storage.storage import Storage


class S3Storage(Storage):
    _client: Any  # boto3 s3 client
    _config: S3StorageConfig

    def __init__(self, config: S3StorageConfig):
        self._client = boto3.client("s3", config=BotoConfig(region_name=config.region))
        self._config = config

    def put(self, key: str, value: bytes) -> None:
        print(f"writing to s3: {key}")
        self._client.put_object(Bucket=self._config.bucket, Key=key, Body=value)

    def get(self, key: str) -> bytes:
        print(f"reading from s3: {key}")
        return self._client.get_object(Bucket=self._config.bucket, Key=key)["Body"].read()
