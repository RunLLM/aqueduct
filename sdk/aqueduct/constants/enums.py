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
    CASCADE = "cascade"


class ServiceType(str, Enum, metaclass=MetaEnum):
    POSTGRES = "Postgres"
    SNOWFLAKE = "Snowflake"
    MYSQL = "MySQL"
    REDSHIFT = "Redshift"
    MARIADB = "MariaDB"
    SQLSERVER = "SQL Server"
    BIGQUERY = "BigQuery"
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
    CONDA = "Conda"
    AQUEDUCT_ENGINE = "Aqueduct"
    DATABRICKS = "Databricks"
    EMAIL = "Email"
    SLACK = "Slack"
    SPARK = "Spark"
    AWS = "AWS"
    ECR = "ECR"
    FILESYSTEM = "Filesystem"
    GAR = "GAR"


class CloudProviderType(str, Enum, metaclass=MetaEnum):
    AWS = "AWS"
    GCP = "GCP"


class RelationalDBServices(str, Enum, metaclass=MetaEnum):
    """Must match the corresponding entries in `ServiceType` exactly."""

    POSTGRES = "Postgres"
    SNOWFLAKE = "Snowflake"
    MYSQL = "MySQL"
    REDSHIFT = "Redshift"
    MARIADB = "MariaDB"
    SQLSERVER = "SQL Server"
    BIGQUERY = "BigQuery"
    SQLITE = "SQLite"
    ATHENA = "Athena"


class ExecutionStatus(str, Enum, metaclass=MetaEnum):
    UNKNOWN = "unknown"
    SUCCEEDED = "succeeded"
    RUNNING = "running"
    FAILED = "failed"
    PENDING = "pending"
    REGISTERED = "registered"
    CANCELED = "canceled"


class NotificationLogLevel(str, Enum, metaclass=MetaEnum):
    SUCCESS = "success"
    WARNING = "warning"
    ERROR = "error"
    INFO = "info"
    NEUTRAL = "neutral"


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


class S3SerializationType(str, Enum, metaclass=MetaEnum):
    CSV_TABLE = "csv_table"
    JSON_TABLE = "json_table"
    PARQUET_TABLE = "parquet_table"
    JSON = "json"
    BYTES = "bytes"
    IMAGE = "image"
    PICKLE = "pickle"


class SerializationType(str, Enum, metaclass=MetaEnum):
    TABLE = "table"
    BSON_TABLE = "bson_table"
    JSON = "json"
    PICKLE = "pickle"
    IMAGE = "image"
    STRING = "string"
    BYTES = "bytes"
    TF_KERAS = "tensorflow-keras-model"


class ExecutionMode(str, Enum, metaclass=MetaEnum):
    EAGER = "eager"
    LAZY = "lazy"


class RuntimeType(str, Enum, metaclass=MetaEnum):
    AQUEDUCT = "aqueduct"
    AQUEDUCT_CONDA = "aqueduct_conda"
    AIRFLOW = "airflow"
    K8S = "k8s"
    LAMBDA = "lambda"
    DATABRICKS = "databricks"
    SPARK = "spark"


class NotificationLevel(Enum, metaclass=MetaEnum):
    SUCCESS = "success"
    WARNING = "warning"
    ERROR = "error"


class SparkRuntimeType(str, Enum, metaclass=MetaEnum):
    DATABRICKS = "databricks"
    SPARK = "spark"


# The lifecycle of a cluster goes from Creating -> Active -> Terminating -> Terminated.
# If an Active cluster receives an config update request, it goes from Active -> Updating -> Active.
class K8sClusterStatusType(str, Enum, metaclass=MetaEnum):
    # The cluster is being created (transitioning from Terminated to Active status).
    CREATING = "Creating"
    # The cluster has been active, and it received a config update request and it's
    # updating to the new config. After the update completes, it will be in the Active status again.
    UPDATING = "Updating"
    # The cluster is currently active and ready to serve requests.
    ACTIVE = "Active"
    # The cluster is being deleted.
    TERMINATING = "Terminating"
    # The cluster is terminated.
    TERMINATED = "Terminated"


class LocalDataTableFormat(str, Enum, metaclass=MetaEnum):
    CSV = "CSV"
    JSON = "JSON"
    PARQUET = "Parquet"


class LocalDataSerializationType(str, Enum, metaclass=MetaEnum):
    CSV_TABLE = "csv_table"
    JSON_TABLE = "json_table"
    PARQUET_TABLE = "parquet_table"
    IMAGE = "image"
    JSON = "json"
    PICKLE = "pickle"
    STRING = "string"
    BYTES = "bytes"
    TF_KERAS = "tensorflow-keras-model"


class K8sClusterActionType(str, Enum, metaclass=MetaEnum):
    CREATE = "create"
    UPDATE = "update"
    DELETE = "delete"
    FORCE_DELETE = "force-delete"


# These are keys for resources that users can customize for operators.
# They are used as dict keys, eg @op(resources={'num_cpus': 1, ...})
class CustomizableResourceType(str, Enum, metaclass=MetaEnum):
    NUM_CPUS = "num_cpus"
    MEMORY = "memory"
    GPU_RESOURCE_NAME = "gpu_resource_name"
    CUDA_VERSION = "cuda_version"
    USE_LLM = "use_llm"
