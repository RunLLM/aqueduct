from typing import Optional

from pydantic import BaseModel


class GlobalConfig(BaseModel):
    """Defines the fields that are globally configurable with `aq.global_config()`."""

    lazy: bool
    engine: Optional[str]


GLOBAL_LAZY_KEY = "lazy"
GLOBAL_ENGINE_KEY = "engine"
__GLOBAL_CONFIG__ = GlobalConfig(lazy=False)
