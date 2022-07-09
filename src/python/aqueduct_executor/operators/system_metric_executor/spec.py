import json
from typing import List

from pydantic import BaseModel, parse_obj_as

try:
    from typing import Literal
except ImportError:
    # Python 3.7 does not support typing.Literal
    from typing_extensions import Literal  # type: ignore

from aqueduct_executor.operators.utils import enums
from aqueduct_executor.operators.utils.storage import config


class SystemMetricSpec(BaseModel):
    name: str
    type: Literal[enums.JobType.SYSTEM_METRIC]
    storage_config: config.StorageConfig
    metadata_path: str
    metric_name: str
    input_metadata_paths: List[str]
    output_content_path: str
    output_metadata_path: str


def parse_spec(spec_json: bytes) -> SystemMetricSpec:
    """
    Parses a JSON string into a FunctionSpec.
    """
    data = json.loads(spec_json)
    return parse_obj_as(SystemMetricSpec, data)
