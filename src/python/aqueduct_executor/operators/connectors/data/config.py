from enum import Enum
from typing import Optional, Union

from aqueduct_executor.operators.connectors.data import models
from aqueduct_executor.operators.utils.enums import MetaEnum
from pydantic import Field


class BigQueryConfig(models.BaseConfig):
    project_id: str
    service_account_credentials: str


class MySqlConfig(models.BaseConfig):
    username: str
    password: str
    database: str
    host: str
    port: str


class PostgresConfig(models.BaseConfig):
    username: str
    password: str
    database: str
    host: str
    port: Optional[str] = "5432"


class S3CredentialType(str, Enum, metaclass=MetaEnum):
    ACCESS_KEY = "access_key"
    CONFIG_FILE_PATH = "config_file_path"
    CONFIG_FILE_CONTENT = "config_file_content"


class S3Config(models.BaseConfig):
    # default type to ACCESS_KEY mainly for backward compatibility
    type: S3CredentialType = S3CredentialType.ACCESS_KEY

    # Access key credentials
    access_key_id: str = ""
    secret_access_key: str = ""

    # Config credentials
    config_file_path: str = ""
    config_file_content: str = ""
    config_file_profile: str = ""

    bucket: str = ""

    region: str = ""
    use_as_storage: str = ""


class SnowflakeConfig(models.BaseConfig):
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


class SqlServerConfig(models.BaseConfig):
    username: str
    password: str
    database: str
    host: str
    port: str


class SqliteConfig(models.BaseConfig):
    database: str


Config = Union[
    BigQueryConfig,
    MySqlConfig,
    PostgresConfig,
    S3Config,
    SnowflakeConfig,
    SqlServerConfig,
    SqliteConfig,
]
