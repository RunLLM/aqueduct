import uuid

from pydantic import BaseModel
from typing import Any, Dict, Optional


class ArtifactSpec(BaseModel):
    table: Optional[Dict[Any, Any]]
    float: Optional[Dict[Any, Any]]
    bool: Optional[Dict[Any, Any]]
    jsonable: Optional[Dict[Any, Any]]


class Artifact(BaseModel):
    id: uuid.UUID
    name: str
    spec: ArtifactSpec
