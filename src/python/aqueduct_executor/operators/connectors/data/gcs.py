import os
import uuid
from typing import Any, List

from aqueduct_executor.operators.connectors.data import connector
from aqueduct_executor.operators.connectors.data.config import GCSConfig
from aqueduct_executor.operators.utils.enums import ArtifactType
from aqueduct_executor.operators.utils.saved_object_delete import SavedObjectDelete
from google.cloud import storage

_CREDENTIALS_ENV_VAR = "GCS_CREDENTIALS"


class GCSConnector(connector.DataConnector):
    _client: Any  # GCS client
    _config: GCSConfig

    def __init__(self, config: GCSConfig):
        if _CREDENTIALS_ENV_VAR in os.environ in os.environ:
            # GCS credentials were provided via env variables instead of a filepath
            temp_path = os.path.join(os.getcwd(), str(uuid.uuid4()))
            with open(temp_path, "w") as f:
                f.write(os.environ[_CREDENTIALS_ENV_VAR])
            config.credentials_path = temp_path

        self._client = storage.Client.from_service_account_json(config.credentials_path)
        self._config = config

    def __del__(self):
        # Try to clean up temp credentials file
        os.remove(self._config.credentials_path)

    def authenticate(self) -> None:
        self._client.list_buckets()

    def discover(self) -> List[str]:
        raise Exception("Discover is not supported for GCS.")

    def extract(self, params: Any) -> Any:
        raise Exception("Extract is not currently supported for GCS.")

    def load(self, params: Any, data: Any, artifact_type: ArtifactType) -> None:
        raise Exception("Load is not currently supported for GCS.")

    def delete(self, objects: Any) -> List[SavedObjectDelete]:
        raise Exception("Delete is not currently supported for GCS.")

    def _delete_object(self, name, context) -> None:
        raise Exception("Delete helper is not implemented for GCS.")
