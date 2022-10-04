import json

from aqueduct_executor.operators.utils.storage import config
from pydantic import BaseModel, parse_obj_as


class MigrationSpec(BaseModel):
    param_val: str
    param_type: str
    op: str


def parse_spec(spec_json: bytes) -> MigrationSpec:
    """
    Parses a JSON string into a MigrationSpec.
    """
    data = json.loads(spec_json)
    return parse_obj_as(MigrationSpec, data)
