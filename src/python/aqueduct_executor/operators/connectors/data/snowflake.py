from typing import Dict

import pandas as pd
from aqueduct_executor.operators.connectors.data import config, relational, utils, load
from aqueduct_executor.operators.utils.enums import ArtifactType
from sqlalchemy import create_engine, engine
from sqlalchemy.types import VARCHAR


class SnowflakeConnector(relational.RelationalConnector):
    def __init__(self, config: config.SnowflakeConfig):
        conn_engine = create_snowflake_engine(config)
        super().__init__(conn_engine)


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


def _map_object_dtype_to_varchar(self, df: pd.DataFrame) -> Dict[str, VARCHAR]:
    col_to_type = {}
    for col in df.select_dtypes(include=["object"]):
        col_to_type[col] = VARCHAR()
    return col_to_type


def load(
        self, params: load.RelationalParams, df: pd.DataFrame, artifact_type: ArtifactType
    ) -> None:
        if artifact_type != ArtifactType.TABLE:
            raise Exception("The data being loaded must be of type table, found %s" % artifact_type)

        # Map of only string columns to their SQL type
        # If a column is not in this map, we rely on Panda's default mapping
        col_to_type = self._map_object_dtype_to_varchar(df)

        # NOTE (saurav): df._to_sql has known performance issues. Using `method="multi"` helps incrementally,
        # since pandas will pass multiple rows in a single INSERT. If this still remains an issue, we can pass in a
        # callable function for `method` that does bulk loading.
        # See: https://pandas.pydata.org/docs/user_guide/io.html#io-sql-method
        df.to_sql(
            params.table,
            con=self.engine,
            if_exists=params.update_mode.value,
            index=False,
            dtype=col_to_type,
            method="multi",
        )