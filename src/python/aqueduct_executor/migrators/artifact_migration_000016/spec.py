import json

from aqueduct_executor.operators.utils.storage import config
from pydantic import BaseModel, parse_obj_as


class MigrationSpec(BaseModel):
    artifact_type: str
    storage_config: config.StorageConfig
    metadata_path: str
    content_path: str


def parse_spec(spec_json: bytes) -> MigrationSpec:
    """
    Parses a JSON string into a FunctionSpec.
    """
    data = json.loads(spec_json)
    return parse_obj_as(MigrationSpec, data)
