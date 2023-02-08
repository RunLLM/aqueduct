from enum import Enum
from typing import Any, Dict, Optional, Union, cast

from aqueduct.constants.enums import MetaEnum, ServiceType
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
    username: str
    password: str
    account_identifier: str
    database: str
    warehouse: str
    db_schema: Optional[str] = Field("public", alias="schema")  # schema is a Pydantic keyword

    class Config:
        # Ensures that Pydantic parses JSON keys named "schema" or "db_schema" to
        # the `db_schema` field
        allow_population_by_field_name = True


class SqlServerConfig(BaseConnectionConfig):
    username: str
    password: str
    database: str
    host: str
    port: str


class SQLiteConfig(BaseConnectionConfig):
    database: str


IntegrationConfig = Union[
    BigQueryConfig,
    MySQLConfig,
    MongoDBConfig,
    PostgresConfig,
    S3Config,
    AthenaConfig,
    SnowflakeConfig,
    SqlServerConfig,
    SQLiteConfig,
]


def convert_dict_to_integration_connect_config(
    service: ServiceType, config_dict: Dict[str, str]
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
    raise InternalAqueductError("Unexpected Service Type: %s" % service)


def prepare_integration_config(
    service: ServiceType, config: IntegrationConfig
) -> IntegrationConfig:
    """Prepares the IntegrationConfig object to be sent to the backend
    as part of a request to connect a new integration.
    """
    if service == ServiceType.BIGQUERY:
        return _prepare_big_query_config(cast(BigQueryConfig, config))
    return config


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
