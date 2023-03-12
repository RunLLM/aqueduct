from pydantic import BaseModel
from typing import Optional, Union
from aqueduct.constants.enums import (
    ArtifactType,
    LocalDataTableFormat,
)
class LocalData(BaseModel):
    path : str
    artifact_type : ArtifactType
    format : Optional[LocalDataTableFormat]
