from enum import Enum, EnumMeta
from typing import Any, Iterable, cast

"""
All these enums can be replaced with protobufs for consistency with the backend.
"""


class MetaEnum(EnumMeta):
    """Allows to very easily check if strings are present in the enum, without a helper.

    Eg.
        if "Postgres" in ServiceType:
            ...
    """

    def __contains__(cls, item: Any) -> Any:
        return item in [v.value for v in cast(Iterable[Enum], cls.__members__.values())]


class FunctionType(str, Enum, metaclass=MetaEnum):
    FILE = "file"
    CODE = "code"
    GITHUB = "github"
    BUILTIN = "built_in"


class FunctionGranularity(str, Enum, metaclass=MetaEnum):
    TABLE = "table"
    ROW = "row"


class CheckSeverity(str, Enum, metaclass=MetaEnum):
    """An ERROR severity will fail the flow."""

    WARNING = "warning"
    ERROR = "error"


class OperatorType(Enum, metaclass=MetaEnum):
    EXTRACT = "extract"
    LOAD = "load"
    FUNCTION = "function"
    METRIC = "metric"
    CHECK = "check"
    PARAM = "param"
    SYSTEM_METRIC = "system_metric"


class TriggerType(Enum, metaclass=MetaEnum):
    MANUAL = "manual"
    PERIODIC = "periodic"


class ServiceType(str, Enum, metaclass=MetaEnum):
    POSTGRES = "Postgres"
    SNOWFLAKE = "Snowflake"
    MYSQL = "MySQL"
    REDSHIFT = "Redshift"
    MARIADB = "MariaDB"
    SQLSERVER = "SQL Server"
    BIGQUERY = "BigQuery"
    AQUEDUCTDEMO = "Aqueduct Demo"
    GITHUB = "Github"
    SALESFORCE = "Salesforce"
    GOOGLE_SHEETS = "Google Sheets"
    S3 = "S3"
    ATHENA = "Athena"
    SQLITE = "SQLite"
    AIRFLOW = "Airflow"
    K8S = "Kubernetes"
    GCS = "GCS"
    LAMBDA = "Lambda"
    MONGO_DB = "MongoDB"


class RelationalDBServices(str, Enum, metaclass=MetaEnum):
    POSTGRES = "Postgres"
    SNOWFLAKE = "Snowflake"
    MYSQL = "MySQL"
    REDSHIFT = "Redshift"
    MARIADB = "MariaDB"
    SQLSERVER = "SQL Server"
    BIGQUERY = "BigQuery"
    AQUEDUCTDEMO = "Aqueduct Demo"
    SQLITE = "SQLite"
    ATHENA = "Athena"


class ExecutionStatus(str, Enum, metaclass=MetaEnum):
    UNKNOWN = "unknown"
    SUCCEEDED = "succeeded"
    FAILED = "failed"
    PENDING = "pending"
    REGISTERED = "registered"
    CANCELED = "canceled"


class FailureType(Enum, metaclass=MetaEnum):
    SYSTEM = 1
    USER_FATAL = 2
    # For failures that don't stop execution.
    # Eg. check operator with WARNING severity fails.
    USER_NON_FATAL = 3


class SalesforceExtractType(str, Enum, metaclass=MetaEnum):
    SEARCH = "search"
    QUERY = "query"


class S3TableFormat(str, Enum, metaclass=MetaEnum):
    CSV = "CSV"
    JSON = "JSON"
    PARQUET = "Parquet"


class LoadUpdateMode(str, Enum, metaclass=MetaEnum):
    APPEND = "append"
    REPLACE = "replace"
    FAIL = "fail"


class GoogleSheetsSaveMode(str, Enum, metaclass=MetaEnum):
    OVERWRITE = "overwrite"
    NEWSHEET = "newsheet"
    CREATE = "create"


class GithubRepoConfigContentType(str, Enum, metaclass=MetaEnum):
    """Github repo config (.aqconfig) content type."""

    OPERATOR = "operator"
    QUERY = "query"


# This is only for displaying the DAG.
class DisplayNodeType(str, Enum, metaclass=MetaEnum):
    OPERATOR = "OPERATOR"
    ARTIFACT = "ARTIFACT"


class ArtifactType(str, Enum, metaclass=MetaEnum):
    UNTYPED = "untyped"
    STRING = "string"
    BOOL = "boolean"
    NUMERIC = "numeric"
    DICT = "dictionary"
    TUPLE = "tuple"
    LIST = "list"
    TABLE = "table"
    JSON = "json"
    BYTES = "bytes"
    IMAGE = "image"  # corresponds to PIL.Image.Image type
    PICKLABLE = "picklable"
    TF_KERAS = "tensorflow-keras-model"


class SerializationType(str, Enum, metaclass=MetaEnum):
    TABLE = "table"
    JSON = "json"
    PICKLE = "pickle"
    IMAGE = "image"
    STRING = "string"
    BYTES = "bytes"
    TF_KERAS = "tensorflow-keras-model"


class ExecutionMode(str, Enum, metaclass=MetaEnum):
    EAGER = "eager"
    LAZY = "lazy"


class RuntimeType(Enum, metaclass=MetaEnum):
    AQUEDUCT = "aqueduct"
    AIRFLOW = "airflow"
    K8S = "k8s"
    LAMBDA = "lambda"
