from aqueduct.api_client import APIClient
from pydantic import BaseModel


class GlobalConfig(BaseModel):
    """Defines the fields that are globally configurable with `aq.global_config()`."""

    lazy: bool


GLOBAL_LAZY_KEY = "lazy"
__GLOBAL_CONFIG__ = GlobalConfig(lazy=False)

# Initialize a unconfigured api client. It will be configured when the user construct an Aqueduct client.
__GLOBAL_API_CLIENT__ = APIClient()
