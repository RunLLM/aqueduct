import importlib
import json
import os
import shutil
import sys
import tracemalloc
import uuid
from typing import Any, Callable, Dict, List, Tuple

import numpy as np
import pandas as pd
from aqueduct.utils.type_inference import infer_artifact_type
from aqueduct_executor.operators.function_executor import extract_function, get_extract_path
from aqueduct_executor.operators.function_executor.execute import (
    cleanup,
    get_py_import_path,
    import_invoke_method,
    run_helper,
    validate_spec,
)
from aqueduct_executor.operators.function_executor.spec import FunctionSpec
from aqueduct_executor.operators.function_executor.utils import OP_DIR
from aqueduct_executor.operators.spark.utils import read_artifacts_spark, write_artifact_spark
from aqueduct_executor.operators.utils import utils
from aqueduct_executor.operators.utils.enums import (
    ArtifactType,
    CheckSeverity,
    ExecutionStatus,
    FailureType,
    OperatorType,
    SerializationType,
)
from aqueduct_executor.operators.utils.execution import (
    TIP_CHECK_DID_NOT_PASS,
    TIP_NOT_BOOL,
    TIP_NOT_NUMERIC,
    TIP_OP_EXECUTION,
    TIP_UNKNOWN_ERROR,
    ExecFailureException,
    ExecutionState,
    Logs,
    exception_traceback,
)
from aqueduct_executor.operators.utils.storage.parse import parse_storage
from aqueduct_executor.operators.utils.timer import Timer
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
    return run_helper(
        spec=spec, 
        read_func=read_artifacts_spark, 
        write_func=write_artifact_spark, 
        infer_func=infer_artifact_type_spark, 
        spark_session_obj=spark_session_obj,
    )
