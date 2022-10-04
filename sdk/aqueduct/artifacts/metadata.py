import uuid
from typing import Optional

from aqueduct.enums import ArtifactType
from pydantic import BaseModel


class ArtifactMetadata(BaseModel):
    id: uuid.UUID
    name: str
    type: ArtifactType
