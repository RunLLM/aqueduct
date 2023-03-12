from typing import Optional, Union

from aqueduct.constants.enums import ArtifactType, LocalDataTableFormat
from pydantic import BaseModel


class LocalData(BaseModel):
    path: str
    artifact_type: ArtifactType
    format: Optional[LocalDataTableFormat]
