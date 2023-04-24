from typing import Any

from aqueduct.utils.type_inference import infer_artifact_type
from aqueduct_executor.operators.function_executor.execute import execute_function_spec
from aqueduct_executor.operators.function_executor.spec import FunctionSpec
from aqueduct_executor.operators.spark.utils import read_artifacts_spark, write_artifact_spark
from aqueduct_executor.operators.utils.enums import ArtifactType
from pyspark.sql import SparkSession, dataframe


def infer_artifact_type_spark(value: Any) -> Any:
    if isinstance(value, dataframe.DataFrame):
        return ArtifactType.TABLE
    else:
        return infer_artifact_type(value)


def run(spec: FunctionSpec, spark_session_obj: SparkSession) -> None:
    """
    Executes a function operator.
    """
    return execute_function_spec(
        spec=spec,
        read_artifacts_func=read_artifacts_spark,
        write_artifact_func=write_artifact_spark,
        infer_type_func=infer_artifact_type_spark,
        spark_session_obj=spark_session_obj,
    )
