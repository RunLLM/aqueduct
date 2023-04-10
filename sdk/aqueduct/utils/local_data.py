import os
from typing import Optional

from aqueduct.constants.enums import ArtifactType, LocalDataTableFormat

from ..error import InvalidUserArgumentException


def _convert_to_local_data_table_format(format: Optional[str]) -> Optional[LocalDataTableFormat]:
    """A simple string -> enum conversion. Returns None if no format provided."""
    if format is None:
        return None

    lowercased_format = format.lower()
    if lowercased_format == LocalDataTableFormat.CSV.value.lower():
        format_enum = LocalDataTableFormat.CSV
    elif lowercased_format == LocalDataTableFormat.JSON.value.lower():
        format_enum = LocalDataTableFormat.JSON
    elif lowercased_format == LocalDataTableFormat.PARQUET.value.lower():
        format_enum = LocalDataTableFormat.PARQUET
    else:
        raise InvalidUserArgumentException("Unsupported local file format `%s`." % format)
    return format_enum


def validate_local_data(
    path: str, artifact_type: Optional[ArtifactType], format: Optional[str]
) -> None:
    """Validate Local Data on its file path and types."""
    file_path = path
    artifact_type = artifact_type
    file_format = format

    if artifact_type is None:
        raise InvalidUserArgumentException(
            "Specify artifact type in `as_type` field in `create_param` to use local data. "
        )

    if not os.path.exists(file_path):
        raise InvalidUserArgumentException(
            "Given path file '%s' to local data does not exist." % file_path
        )

    if artifact_type == ArtifactType.TABLE and file_format is None:
        raise InvalidUserArgumentException(
            "Specify format in order to use local data as TableArtifact."
        )
