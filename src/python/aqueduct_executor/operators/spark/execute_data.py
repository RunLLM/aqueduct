import sys
from typing import Any

from aqueduct_executor.operators.connectors.data import common, config, connector
from aqueduct_executor.operators.connectors.data.execute import execute_data_spec
from aqueduct_executor.operators.connectors.data.spec import Spec
from aqueduct_executor.operators.spark.utils import read_artifacts_spark, write_artifact_spark
from aqueduct_executor.operators.utils.exceptions import (
    MissingConnectorDependencyException,
    UnsupportedConnectorExecption,
)
from pyspark.sql import SparkSession


def run(spec: Spec, spark_session_obj: SparkSession) -> None:
    """
    Runs one of the following connector operations:
    - authenticate
    - extract
    - load
    - load-table
    - delete-saved-objects
    - discover

    Arguments:
    - spec: The spec provided for this operator.
    - spark_session_obj: The SparkSession
    """
    return execute_data_spec(
        spec=spec,
        read_artifacts_func=read_artifacts_spark,
        write_artifact_func=write_artifact_spark,
        setup_connector_func=setup_connector_spark,
        is_spark=True,
        spark_session_obj=spark_session_obj,
    )


def setup_connector_spark(
    connector_name: common.Name, connector_config: config.Config
) -> connector.DataConnector:
    """
    Sets up the Spark Connectors. We currently support the following resources with Spark:
    - S3
    - Snowflake
    """
    # prevent isort from moving around type: ignore comments which will cause mypy issues.
    # isort: off
    if connector_name == common.Name.SNOWFLAKE:
        try:
            from pyspark.sql import SparkSession
            from snowflake import sqlalchemy
        except:
            raise MissingConnectorDependencyException(
                "Unable to initialize the Spark Snowflake connector. Have you run `aqueduct install spark-snowflake`?"
            )

        from aqueduct_executor.operators.connectors.data.spark.snowflake import (
            SparkSnowflakeConnector as OpConnector,
        )
    elif connector_name == common.Name.S3:
        try:
            import pyarrow
        except:
            raise MissingConnectorDependencyException(
                "Unable to initialize the Spark S3 connector. Have you run `aqueduct install s3`?"
            )

        from aqueduct_executor.operators.connectors.data.spark.s3 import (  # type: ignore
            SparkS3Connector as OpConnector,
        )
    else:
        raise UnsupportedConnectorExecption(
            "Unable to initialize connector. This connector is not yet supported for Aqueduct on Spark."
        )
    # isort: on
    return OpConnector(config=connector_config)  # type: ignore
