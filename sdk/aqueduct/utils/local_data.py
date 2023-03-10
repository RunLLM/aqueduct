import os
from aqueduct.constants.enums import ArtifactType,LocalDataTableFormat
from typing import Any, DefaultDict, Dict, List, Optional, Union
from ..error import InvalidUserArgumentException

class LocalData:
    path : str
    as_type : Union[ArtifactType,LocalDataTableFormat]
    format : Optional[LocalDataTableFormat]

def Local_Data(path : str, artifact_type : ArtifactType, format: Optional[str] = None,) -> LocalData:
    """Identify the local data which can be passed in as a parameter.

    Args:
        path:
            The path to the data.
        as_type:
                The expected type of the data. The `ArtifactType` class in `enums.py` contains all
                supported types, except for ArtifactType.UNTYPED. 
        format:
                If the artifact type is ArtifactType.TABLE, the user has to specify the table format.
                We currently support JSON, CSV, and Parquet.
    
    Returns
        A `LocalData` object which can be used to create a parameter.
    """
    if path[-1] == "/":
        raise InvalidUserArgumentException(
            "Path argument is only applicable to file type, found directory instead."
        )
    
    if format and artifact_type != ArtifactType.TABLE:
        raise InvalidUserArgumentException(
            "Format argument is only applicable to table artifact type, found %s instead."
            % artifact_type
        )
    
    format_enum = _convert_to_local_data_table_format(format)
    return LocalData(path,artifact_type,format_enum) 


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