from enum import Enum, EnumMeta
from typing import Any, Iterable, cast


class MetaEnum(EnumMeta):
    """Allows to very easily check if strings are present in the enum, without a helper.
    Eg.
        if "Postgres" in ServiceType:
            ...
    """

    def __contains__(cls, item: Any) -> Any:
        return item in [v.value for v in cast(Iterable[Enum], cls.__members__.values())]


class OperatorType(str, Enum, metaclass=MetaEnum):
    FUNCTION = "function"
    METRIC = "metric"
    CHECK = "check"
    EXTRACT = "extract"
    LOAD = "load"
    PARAM = "param"
    SYSTEM_METRIC = "system_metric"


class CheckSeverityLevel(str, Enum, metaclass=MetaEnum):
    ERROR = "error"
    WARNING = "warning"


class ExecutionStatus(str, Enum, metaclass=MetaEnum):
    UNKNOWN = "unknown"
    PENDING = "pending"
    SUCCEEDED = "succeeded"
    FAILED = "failed"


class FailureType(Enum, metaclass=MetaEnum):
    SYSTEM = 1
    USER_FATAL = 2
    # For failures that don't stop execution.
    # Eg. check operator with WARNING severity fails.
    USER_NON_FATAL = 3


class JobType(str, Enum, metaclass=MetaEnum):
    FUNCTION = "function"
    AUTHENTICATE = "authenticate"
    EXTRACT = "extract"
    LOAD = "load"
    LOADTABLE = "load-table"
    DELETESAVEDOBJECTS = "delete-saved-objects"
    DISCOVER = "discover"
    PARAM = "param"
    SYSTEM_METRIC = "system_metric"
    COMPILE_AIRFLOW = "compile_airflow"


class ArtifactType(str, Enum, metaclass=MetaEnum):
    STRING = "string"
    BOOL = "boolean"
    NUMERIC = "numeric"
    DICT = "dictionary"
    TUPLE = "tuple"
    TABLE = "table"
    JSON = "json"
    BYTES = "bytes"
    IMAGE = "image"
    PICKLABLE = "picklable"


class SerializationType(str, Enum, metaclass=MetaEnum):
    TABLE = "table"
    JSON = "json"
    PICKLE = "pickle"
    IMAGE = "image"
    STRING = "string"
    BYTES = "bytes"


artifact_to_serialization = {
    ArtifactType.STRING: [SerializationType.STRING],
    ArtifactType.BOOL: [SerializationType.JSON],
    ArtifactType.NUMERIC: [SerializationType.JSON],
    ArtifactType.DICT: [SerializationType.JSON, SerializationType.PICKLE],
    ArtifactType.TUPLE: [SerializationType.JSON, SerializationType.PICKLE],
    ArtifactType.TABLE: [SerializationType.TABLE],
    ArtifactType.JSON: [SerializationType.STRING],
    ArtifactType.BYTES: [SerializationType.BYTES],
    ArtifactType.IMAGE: [SerializationType.IMAGE],
    ArtifactType.PICKLABLE: [SerializationType.PICKLE],
}
