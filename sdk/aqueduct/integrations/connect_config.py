import json
from enum import Enum
from typing import Any, Dict, List, Optional, Union, cast

from aqueduct.constants.enums import MetaEnum, NotificationLevel, ServiceType
from aqueduct.error import InternalAqueductError, InvalidUserArgumentException
from pydantic import BaseModel, Extra, Field

"""Copied mostly over from `aqueduct_executor/operators/connectors/data/config.py` for now, please keep them in sync."""


class BaseConnectionConfig(BaseModel):
    """
    BaseConfig defines the Pydantic Config shared by all connector Config's, e.g.
    postgres.Config, mysql.Config, etc.
    """

    class Config:
        extra = Extra.forbid


class BigQueryConfig(BaseConnectionConfig):
    """
    BigQueryConfig defines the Pydantic Config for a BigQuery integration.
    One of the following between `service_account_credentials` and
    `service_account_credentials_path` must be specified. If `service_account_credentials_path`
    is specified, it takes priority.
    """

    project_id: str
    service_account_credentials: Optional[str] = None
    service_account_credentials_path: Optional[str] = None

    def json(self, **kwargs: Any) -> Any:
        """Overrides default JSON serialization to ensure that `service_account_credentials_path`
        is not passed along to the backend.
        """
        return super().json(exclude={"service_account_credentials_path"}, **kwargs)


class MySQLConfig(BaseConnectionConfig):
    username: str
    password: str
    database: str
    host: str
    port: str


class MongoDBConfig(BaseConnectionConfig):
    auth_uri: str
    database: str


class PostgresConfig(BaseConnectionConfig):
    username: str
    password: str
    database: str
    host: str
    # Postgres runs on port 5432 by default
    port: Optional[str] = "5432"


class RedshiftConfig(PostgresConfig):
    # Redshift runs on port 5439 by default
    port: Optional[str] = "5439"


class AWSCredentialType(str, Enum, metaclass=MetaEnum):
    ACCESS_KEY = "access_key"
    CONFIG_FILE_PATH = "config_file_path"
    CONFIG_FILE_CONTENT = "config_file_content"


class S3Config(BaseConnectionConfig):
    # default type to ACCESS_KEY mainly for backward compatibility
    type: AWSCredentialType = AWSCredentialType.ACCESS_KEY

    # Access key credentials
    access_key_id: str = ""
    secret_access_key: str = ""

    # Config credentials
    config_file_path: str = ""
    config_file_profile: str = ""

    bucket: str
    region: str

    # When connecting a new integration, we allow both leading or trailing slashes here.
    # The path will be sanitized before being stored in the database.
    root_dir: str = ""

    use_as_storage: str = "false"


class GCSConfig(BaseConnectionConfig):
    bucket: str
    service_account_credentials: str
    use_as_storage: str = "false"


class AthenaConfig(BaseConnectionConfig):
    # default type to ACCESS_KEY mainly for backward compatibility
    type: AWSCredentialType = AWSCredentialType.ACCESS_KEY

    # Access key credentials
    access_key_id: str = ""
    secret_access_key: str = ""
    region: str = ""

    # Config credentials
    config_file_path: str = ""
    config_file_content: str = ""
    config_file_profile: str = ""

    database: str = ""
    output_location: str = ""


class SnowflakeConfig(BaseConnectionConfig):
    """Must be dumped to JSON with the `exclude_none` and `by_alias` flags."""

    username: str
    password: str
    account_identifier: str
    database: str
    warehouse: str
    db_schema: Optional[str] = Field("public", alias="schema")  # schema is a Pydantic keyword
    role: Optional[str] = None  # we must exclude this field if None when dumping to json.

    class Config:
        # Ensures that Pydantic parses JSON keys named "schema" or "db_schema" to
        # the `db_schema` field. This is only for converting dict -> pydantic object.
        # When dumping this config to a dictionary, be sure to use `SnowflakeConfig.json(by_alias=True)`.
        allow_population_by_field_name = True


class SqlServerConfig(BaseConnectionConfig):
    username: str
    password: str
    database: str
    host: str
    port: str


class SQLiteConfig(BaseConnectionConfig):
    database: str


class SlackConfig(BaseConnectionConfig):
    token: str
    channels: List[str]
    level: Optional[NotificationLevel] = None
    enabled: bool


class _SlackConfigWithStringField(BaseConnectionConfig):
    token: str
    channels_serialized: str
    level: str
    enabled: str


class DynamicK8sConfig(BaseConnectionConfig):
    # How long (in seconds) does the cluster need to remain idle before it is deleted.
    keepalive: Optional[Union[str, int]]
    # The EC2 instance type of the CPU node group. See https://aws.amazon.com/ec2/instance-types/
    # for the node types available.
    cpu_node_type: Optional[str]
    # The EC2 instance type of the GPU node group. See https://aws.amazon.com/ec2/instance-types/
    # for the node types available.
    gpu_node_type: Optional[str]
    # Minimum number of nodes in the CPU node group. The cluster autoscaler cannot scale below this number.
    # This is also the initial number of CPU nodes in the cluster.
    min_cpu_node: Optional[Union[str, int]]
    # Maximum number of nodes in the CPU node group. The cluster autoscaler cannot scale above this number.
    max_cpu_node: Optional[Union[str, int]]
    # Minimum number of nodes in the GPU node group. The cluster autoscaler cannot scale below this number.
    # This is also the initial number of GPU nodes in the cluster.
    min_gpu_node: Optional[Union[str, int]]
    # Maximum number of nodes in the GPU node group. The cluster autoscaler cannot scale above this number.
    max_gpu_node: Optional[Union[str, int]]

    # This converts all int fields to string during json serialization. We need to do this becasue our
    # backend assumes all config fields must be string.
    class Config:
        json_encoders = {int: str}


class AWSConfig(BaseConnectionConfig):
    # Either 1) all of access_key_id, secret_access_key, region, or 2) both config_file_path and
    # config_file_profile need to be specified. Any other cases will be rejected by the server's
    # config validation process.
    access_key_id: str = ""
    secret_access_key: str = ""
    region: str = ""
    config_file_path: str = ""
    config_file_profile: str = ""
    k8s: Optional[DynamicK8sConfig]


class _AWSConfigWithSerializedConfig(BaseConnectionConfig):
    access_key_id: str = ""
    secret_access_key: str = ""
    region: str = ""
    config_file_path: str = ""
    config_file_profile: str = ""
    k8s_serialized: Optional[
        str
    ]  # this is a json-serialized string of AWSConfig.k8s, which is of type DynamicK8sConfig


class EmailConfig(BaseConnectionConfig):
    user: str
    password: str
    host: str
    port: int
    targets: List[str]
    level: Optional[NotificationLevel] = None
    enabled: bool


class _EmailConfigWithStringField(BaseConnectionConfig):
    user: str
    password: str
    host: str
    port: str
    targets_serialized: str
    level: str
    enabled: str


class SparkConfig(BaseConnectionConfig):
    livy_server_url: str


class K8sConfig(BaseConnectionConfig):
    kubeconfig_path: str
    cluster_name: str
    use_same_cluster: str = "false"
    dynamic: str = "false"
    cloud_integration_id: str = ""


IntegrationConfig = Union[
    BigQueryConfig,
    EmailConfig,
    _EmailConfigWithStringField,
    MySQLConfig,
    MongoDBConfig,
    PostgresConfig,
    S3Config,
    AthenaConfig,
    SnowflakeConfig,
    SqlServerConfig,
    SQLiteConfig,
    SlackConfig,
    AWSConfig,
    _AWSConfigWithSerializedConfig,
    _SlackConfigWithStringField,
    SparkConfig,
    K8sConfig,
]


def convert_dict_to_integration_connect_config(
    service: Union[str, ServiceType], config_dict: Dict[str, str]
) -> IntegrationConfig:
    if service == ServiceType.BIGQUERY:
        return BigQueryConfig(**config_dict)
    elif service in [ServiceType.MARIADB, ServiceType.MYSQL]:
        return MySQLConfig(**config_dict)
    elif service == ServiceType.MONGO_DB:
        return MongoDBConfig(**config_dict)
    elif service == ServiceType.POSTGRES:
        return PostgresConfig(**config_dict)
    elif service == ServiceType.S3:
        return S3Config(**config_dict)
    elif service == ServiceType.ATHENA:
        return AthenaConfig(**config_dict)
    elif service == ServiceType.SNOWFLAKE:
        return SnowflakeConfig(**config_dict)
    elif service == ServiceType.SQLSERVER:
        return SqlServerConfig(**config_dict)
    elif service == ServiceType.SQLITE:
        return SQLiteConfig(**config_dict)
    elif service == ServiceType.REDSHIFT:
        return RedshiftConfig(**config_dict)
    elif service == ServiceType.SLACK:
        return SlackConfig(**config_dict)
    elif service == ServiceType.EMAIL:
        return EmailConfig(**config_dict)
    elif service == ServiceType.SPARK:
        return SparkConfig(**config_dict)
    elif service == ServiceType.AWS:
        return AWSConfig(**config_dict)
    elif service == ServiceType.K8S:
        return K8sConfig(**config_dict)
    raise InternalAqueductError("Unexpected Service Type: %s" % service)


def prepare_integration_config(
    service: Union[str, ServiceType], config: IntegrationConfig
) -> IntegrationConfig:
    """Prepares the IntegrationConfig object to be sent to the backend
    as part of a request to connect a new integration.
    """
    if service == ServiceType.BIGQUERY:
        return _prepare_big_query_config(cast(BigQueryConfig, config))

    if service == ServiceType.SLACK:
        return _prepare_slack_config(cast(SlackConfig, config))

    if service == ServiceType.EMAIL:
        return _prepare_email_config(cast(EmailConfig, config))

    if service == ServiceType.AWS:
        return _prepare_aws_config(cast(AWSConfig, config))

    return config


def _prepare_email_config(config: EmailConfig) -> _EmailConfigWithStringField:
    return _EmailConfigWithStringField(
        user=config.user,
        password=config.password,
        host=config.host,
        port=str(config.port),
        targets_serialized=json.dumps(config.targets),
        level=config.level.value if config.level else "",
        enabled="true" if config.enabled else "false",
    )


def _prepare_slack_config(config: SlackConfig) -> _SlackConfigWithStringField:
    return _SlackConfigWithStringField(
        token=config.token,
        channels_serialized=json.dumps(config.channels),
        level=config.level.value if config.level else "",
        enabled="true" if config.enabled else "false",
    )


def _prepare_aws_config(config: AWSConfig) -> _AWSConfigWithSerializedConfig:
    return _AWSConfigWithSerializedConfig(
        access_key_id=config.access_key_id,
        secret_access_key=config.secret_access_key,
        region=config.region,
        config_file_path=config.config_file_path,
        config_file_profile=config.config_file_profile,
        k8s_serialized=(None if config.k8s is None else config.k8s.json(exclude_none=True)),
    )


def _prepare_big_query_config(config: BigQueryConfig) -> BigQueryConfig:
    """Prepares the BigQueryConfig object by reading the service account
    credentials into a string field if the filepath is specified.
    """
    if not config.service_account_credentials and not config.service_account_credentials_path:
        raise InvalidUserArgumentException(
            "At least one of `service_account_credentials` or `service_account_credentials_path` must be set for a BigQueryConfig."
        )

    if not config.service_account_credentials_path:
        return config

    with open(config.service_account_credentials_path, "r") as file:
        credentials = file.read().replace("\n", "")
        config.service_account_credentials = credentials

    return config
