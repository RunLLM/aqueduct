import uuid

from aqueduct.constants.enums import ArtifactType
from pydantic import BaseModel


class ArtifactMetadata(BaseModel):
    id: uuid.UUID
    name: str
    type: ArtifactType
