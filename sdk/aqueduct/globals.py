from typing import Optional

from aqueduct.backend.client import APIClient
from pydantic import BaseModel

from aqueduct.dag.dag import DAG, Metadata


class GlobalConfig(BaseModel):
    """Defines the fields that are globally configurable with `aq.global_config()`."""

    lazy: bool
    engine: Optional[str]


GLOBAL_LAZY_KEY = "lazy"
GLOBAL_ENGINE_KEY = "engine"
__GLOBAL_CONFIG__ = GlobalConfig(lazy=False)

# Initialize an unconfigured api client. It will be configured when the user construct an Aqueduct client.
__GLOBAL_API_CLIENT__ = APIClient()

# Initialize a module-level dag object, to be accessed and modified when the user constructs the flow.
__GLOBAL_DAG__ = DAG(metadata=Metadata())