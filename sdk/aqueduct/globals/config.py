from typing import Optional, Union

from aqueduct.integrations.dynamic_k8s_integration import DynamicK8sIntegration
from pydantic import BaseModel


class GlobalConfig(BaseModel):
    """Defines the fields that are globally configurable with `aq.global_config()`."""

    lazy: bool
    engine: Optional[Union[str, DynamicK8sIntegration]]

    class Config:
        arbitrary_types_allowed = True


GLOBAL_LAZY_KEY = "lazy"
GLOBAL_ENGINE_KEY = "engine"
__GLOBAL_CONFIG__ = GlobalConfig(lazy=False)
