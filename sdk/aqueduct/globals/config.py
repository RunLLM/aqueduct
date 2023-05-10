from typing import Optional, Union

from aqueduct.resources.dynamic_k8s import DynamicK8sResource
from pydantic import BaseModel


class GlobalConfig(BaseModel):
    """Defines the fields that are globally configurable with `aq.global_config()`."""

    lazy: bool
    engine: Optional[Union[str, DynamicK8sResource]]

    class Config:
        arbitrary_types_allowed = True


GLOBAL_LAZY_KEY = "lazy"
GLOBAL_ENGINE_KEY = "engine"
__GLOBAL_CONFIG__ = GlobalConfig(lazy=False)
