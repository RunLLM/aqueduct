from typing import Dict

import pandas as pd
from aqueduct_executor.operators.connectors.data import config, relational, utils
from sqlalchemy import create_engine, engine
from sqlalchemy.types import VARCHAR


class SnowflakeConnector(relational.RelationalConnector):
    def __init__(self, config: config.SnowflakeConfig):
        conn_engine = create_snowflake_engine(config)
        super().__init__(
            conn_engine=conn_engine,
            object_to_varchar_mapper=map_object_dtype_to_varchar,
        )


def create_snowflake_engine(config: config.SnowflakeConfig) -> engine.Engine:
    # Snowflake Dialect:
    # https://github.com/snowflakedb/snowflake-sqlalchemy
    url = "snowflake://{username}:{password}@{account_identifier}/{database}/{schema}?warehouse={warehouse}".format(
        username=config.username,
        password=utils.url_encode(config.password),
        account_identifier=config.account_identifier,
        database=config.database,
        schema=config.db_schema,
        warehouse=config.warehouse,
    )
    return create_engine(url)


def map_object_dtype_to_varchar(df: pd.DataFrame) -> Dict[str, VARCHAR]:
    col_to_type = {}
    for col in df.select_dtypes(include=["object"]):
        col_to_type[col] = VARCHAR()
    return col_to_type
