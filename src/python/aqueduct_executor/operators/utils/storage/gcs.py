import io
import os
import uuid
from typing import Any

from aqueduct_executor.operators.utils.storage.config import GCSStorageConfig
from aqueduct_executor.operators.utils.storage.storage import Storage
from google.cloud import storage

_CREDENTIALS_ENV_VAR = "GCS_CREDENTIALS"


class GCSStorage(Storage):
    _client: Any  # GCS client
    _config: GCSStorageConfig

    def __init__(self, config: GCSStorageConfig):
        if _CREDENTIALS_ENV_VAR in os.environ in os.environ:
            # GCS credentials were provided via env variables instead of a filepath
            temp_path = os.path.join(os.getcwd(), str(uuid.uuid4()))
            with open(temp_path, "w") as f:
                f.write(os.environ[_CREDENTIALS_ENV_VAR])
            config.credentials_path = temp_path

        self._client = storage.Client.from_service_account_json(config.credentials_path)
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
        return blob.download_as_bytes()
