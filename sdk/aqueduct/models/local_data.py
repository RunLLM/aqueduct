import os
from typing import Any, Optional

from aqueduct.constants.enums import ArtifactType
from pydantic import BaseModel

from ..error import InvalidUserArgumentException

MAXIMUM_LOCAL_DATA_SIZE = 1000000


class LocalData(BaseModel):
    path: str
    artifact_type: ArtifactType
    format: Optional[str]

    # Override Pydantic `init` with validation
    def __init__(__pydantic_self__, **data: Any) -> None:
        """Identify the local data which can be passed in as a parameter.

        Args:
            path:
                The path to the data.
            artifact_type:
                    The expected type of the data. Currently LocalData has support for ArtifactType.TABLE
                    and ArtifactType.IMAGE
            format:
                    If the artifact type is ArtifactType.TABLE, the user has to specify the table format.
                    We currently support "JSON", "CSV", and "Parquet".

        Returns
            A configuration object which you can use to reference data in the local filesystem.
        """
        super().__init__(**data)
        __pydantic_self__.validate_local_data()

    def validate_local_data(__pydantic_self__) -> None:
        """Validate LocalData on its file path and types."""
        file_path = __pydantic_self__.path
        artifact_type = __pydantic_self__.artifact_type
        file_format = __pydantic_self__.format

        if not os.path.isfile(file_path):
            raise InvalidUserArgumentException(
                "Given path file '%s' to local data does not exist." % file_path
            )

        if os.path.getsize(file_path) > MAXIMUM_LOCAL_DATA_SIZE:
            raise InvalidUserArgumentException("Currently, the maximum local data size is 1 MB.")
        if artifact_type == ArtifactType.TABLE and file_format is None:
            raise InvalidUserArgumentException(
                "Specify format in order to use local data as TableArtifact."
            )
