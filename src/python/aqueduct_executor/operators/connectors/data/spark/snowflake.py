from typing import Any, Callable, Dict, List, Optional

from aqueduct_executor.operators.connectors.data import common, config, relational, snowflake, utils
from aqueduct_executor.operators.connectors.data import connector, extract, load
from aqueduct_executor.operators.utils.enums import ArtifactType
from pyspark.sql import SparkSession, DataFrame, FloatType, col


class SparkSnowflakeConnector(relational.RelationalConnector):
    def __init__(self, config: config.SnowflakeConfig):
        url = "https://{account_identifier}.snowflakecomputing.com".format(
            account_identifier=config.account_identifier,
        )
        conn_engine = snowflake.create_snowflake_engine(config)

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
        decimals_cols = [c for c in df.columns if 'Decimal' in str(df.schema[c].dataType)]
        #convert all decimals columns to floats
        for col in decimals_cols:
            df = df.withColumn(col, df[col].cast(FloatType()))
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


def _map_to_snowflake_mode(params: load.RelationalParams) -> str:
    """
    Map Aqueduct update modes to Snowflake update modes found here:
    https://docs.snowflake.com/ko/developer-guide/snowpark/reference/python/api/snowflake.snowpark.DataFrameWriter.save_as_table.html
    """
    snowflake_update_mode = "ignore"
    update_mode = params.update_mode.value
    if update_mode == common.UpdateMode.APPEND:
       snowflake_update_mode = "append"
    elif update_mode == common.UpdateMode.REPLACE:
        snowflake_update_mode = "overwrite"
    elif update_mode == common.UpdateMode.FAIL:
        snowflake_update_mode = "errorifexists"
    
    return snowflake_update_mode