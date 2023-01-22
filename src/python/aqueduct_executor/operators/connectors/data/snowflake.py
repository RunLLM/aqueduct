from typing import Any, Callable, Dict, List, Optional

from aqueduct_executor.operators.connectors.data import config, relational, utils
from aqueduct_executor.operators.connectors.data import connector, extract, load
from aqueduct_executor.operators.utils.enums import ArtifactType
from sqlalchemy import create_engine, engine
from pyspark.sql import SparkSession, DataFrame


class SnowflakeConnector(relational.RelationalConnector):
    def __init__(self, config: config.SnowflakeConfig):
        url = "https://{account_identifier}.snowflakecomputing.com".format(
            account_identifier=config.account_identifier,
        )
        conn_engine = _create_engine(config)

        self.snowflake_spark_options = {
            'sfURL': url,
            'sfAccount': config.account_identifier,
            'sfUser': config.username,
            'sfPassword': config.password,
            'sfDatabase': config.database,
            'sfSchema': config.schema,
            'sfWarehouse': config.warehouse,
        }
        super().__init__(conn_engine)
    

    def extract_spark(self, params: extract.RelationalParams, spark_session_obj: SparkSession) -> Any:
        assert params.usable(), "Query is not usable. Did you forget to expand placeholders?"

        df = spark_session_obj.read.format("snowflake").options(**self.snowflake_spark_options).option("query", params.query).load()
        return df


    def load_spark(
            self, params: load.RelationalParams, df: DataFrame, artifact_type: ArtifactType
        ) -> None:
       
        snowflake_update_mode = _map_to_snowflake_mode(params)

        df.write.format("snowflake") \
        .options(**self.snowflake_spark_options) \
        .option("sfSchema", "public") \
        .option("dbtable", params.table) \
        .mode(snowflake_update_mode) \
        .save()


def _create_engine(config: config.SnowflakeConfig) -> engine.Engine:
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


def _map_to_snowflake_mode(params: load.RelationalParams) -> str:
    """
    Map Aqueduct update modes to Snowflake update modes found here:
    https://docs.snowflake.com/ko/developer-guide/snowpark/reference/python/api/snowflake.snowpark.DataFrameWriter.save_as_table.html
    """
    snowflake_update_mode = "ignore"
    if params.update_mode.value == "append":
        snowflake_update_mode = "append"
    elif params.update_mode.value == "fail":
        snowflake_update_mode = "errorifexists"
    elif params.update_mode.value == "replace":
        snowflake_update_mode = "overwrite"
    
    return snowflake_update_mode