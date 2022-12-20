import json
from typing import Any, List

from aqueduct_executor.operators.connectors.data import connector
from aqueduct_executor.operators.connectors.data.config import GCSConfig
from aqueduct_executor.operators.utils.enums import ArtifactType
from aqueduct_executor.operators.utils.saved_object_delete import SavedObjectDelete
from google.cloud import storage
from google.oauth2 import service_account

_CREDENTIALS_ENV_VAR = "GCS_CREDENTIALS"


class GCSConnector(connector.DataConnector):
    _client: Any  # GCS client
    _config: GCSConfig
    _temp_credentials_path: str = ""

    def __init__(self, config: GCSConfig):
        credentials_info = json.loads(config.service_account_credentials)
        credentials = service_account.Credentials.from_service_account_info(credentials_info)
        self._client = storage.Client(credentials=credentials)
        self._config = config

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
