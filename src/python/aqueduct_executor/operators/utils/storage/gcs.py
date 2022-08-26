import io
import json
from typing import Any

from aqueduct_executor.operators.utils.storage.config import GCSStorageConfig
from aqueduct_executor.operators.utils.storage.storage import Storage
from google.cloud import storage
from google.oauth2 import service_account


class GCSStorage(Storage):
    _client: Any  # GCS client
    _config: GCSStorageConfig

    def __init__(self, config: GCSStorageConfig):
        credentials_info = json.loads(config.service_account_credentials)
        credentials = service_account.Credentials.from_service_account_info(credentials_info)
        self._client = storage.Client(credentials=credentials)
        self._config = config

    def put(self, key: str, value: bytes) -> None:
        bucket = self._client.bucket(self._config.bucket)
        blob = bucket.blob(key)

        print(f"writing to gcs: {key}")
        f = io.BytesIO(value)
        blob.upload_from_file(f)

    def get(self, key: str) -> bytes:
        bucket = self._client.bucket(self._config.bucket)
        blob = bucket.blob(key)

        print(f"reading from gcs: {key}")
        return bytes(blob.download_as_bytes())
