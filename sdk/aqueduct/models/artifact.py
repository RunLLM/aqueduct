import uuid

from aqueduct.constants.enums import ArtifactType
from pydantic import BaseModel


class ArtifactMetadata(BaseModel):
    id: uuid.UUID
    name: str
    type: ArtifactType

    # Whether this artifact was given a name explicitly by the user.
    # If true, this artifact name is expected to be unique in the DAG.
    explicitly_named: bool = False

    class Config:
        fields = {"explicitly_named": {"exclude": ...}}
