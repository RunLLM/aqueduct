import uuid
from typing import Any, Dict, Optional

from aqueduct.enums import ArtifactType
from aqueduct.error import AqueductError
from pydantic import BaseModel


class Artifact(BaseModel):
    id: uuid.UUID
    name: str
    type: Optional[ArtifactType]
