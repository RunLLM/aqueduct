import io
import json
from typing import Any, Callable, Dict, List, Optional, Tuple

import pandas as pd
from aqueduct.utils.serialization import (
    DEFAULT_ENCODING,
    artifact_type_to_serialization_type,
    deserialize,
    serialize_val,
)
from aqueduct_executor.operators.utils.enums import (
    ArtifactType,
    ExecutionStatus,
    FailureType,
    SerializationType,
)
from aqueduct_executor.operators.utils.exceptions import MissingInputPathsException
from aqueduct_executor.operators.utils.execution import (
    TIP_UNKNOWN_ERROR,
    Error,
    ExecFailureException,
    ExecutionState,
    Logs,
    exception_traceback,
)
from aqueduct_executor.operators.utils.saved_object_delete import SavedObjectDelete
from aqueduct_executor.operators.utils.storage.storage import Storage
from aqueduct_executor.operators.utils.utils import (
    _METADATA_ARTIFACT_TYPE_KEY,
    _METADATA_PYTHON_TYPE_KEY,
    _METADATA_SCHEMA_KEY,
    _METADATA_SERIALIZATION_TYPE_KEY,
    _METADATA_SYSTEM_METADATA_KEY,
    serialize_val_wrapper,
)
from pyspark.sql import SparkSession


def read_artifacts_spark(
    storage: Storage,
    input_paths: List[str],
    input_metadata_paths: List[str],
    spark_session_obj: SparkSession,
) -> Tuple[List[Any], List[ArtifactType], List[SerializationType]]:
    if len(input_paths) != len(input_metadata_paths):
        raise Exception(
            "Found inconsistent number of input paths (%d) and input metadata paths (%d)"
            % (
                len(input_paths),
                len(input_metadata_paths),
            )
        )

    inputs: List[Any] = []
    artifact_types: List[ArtifactType] = []
    serialization_types: List[SerializationType] = []

    for (input_path, input_metadata_path) in zip(input_paths, input_metadata_paths):

        artifact_metadata = json.loads(storage.get(input_metadata_path).decode(DEFAULT_ENCODING))
        artifact_type = artifact_metadata[_METADATA_ARTIFACT_TYPE_KEY]
        artifact_types.append(artifact_type)

        serialization_type = artifact_metadata[_METADATA_SERIALIZATION_TYPE_KEY]
        serialization_types.append(serialization_type)

        # Check if artifact is of type TABLE. If it is, attempt to read from a temporary view with
        # name of the input_path.
        if artifact_type != ArtifactType.TABLE:
            inputs.append(deserialize(serialization_type, artifact_type, storage.get(input_path)))
        else:
            # read from temp view
            try:
                # global_temp_db = spark_session_obj.conf.get("spark.sql.globalTempDatabase")
                view_path = "global_temp" + "." + convert_path_to_view_name(input_path)
                spark_df = spark_session_obj.read.table(view_path)
                inputs.append(spark_df)
            except Exception as e:
                raise MissingInputPathsException(
                    "Unable to read inputs artifacts from temp view. Exception: %s" % str(e)
                )

    return inputs, artifact_types, serialization_types


def write_artifact_spark(
    storage: Storage,
    artifact_type: ArtifactType,
    # derived_from_bson specifies if the artifact is derived from a bson object
    # and thus requires bson encoding.
    # For now, it only applies to data frames extracted / transformed from Mongo.
    derived_from_bson: bool,
    output_path: Optional[str],
    output_metadata_path: str,
    content: Any,
    system_metadata: Dict[str, str],
    spark_session_obj: SparkSession,
) -> None:
    """The `output_path` can be empty if the contents were already pre-populated (eg. parameter operators)."""
    output_metadata: Dict[str, Any] = {
        _METADATA_SCHEMA_KEY: [],
        _METADATA_SYSTEM_METADATA_KEY: system_metadata,
        _METADATA_ARTIFACT_TYPE_KEY: artifact_type.value,
    }

    if artifact_type == ArtifactType.TABLE:
        output_metadata[_METADATA_SCHEMA_KEY] = [{col[0]: col[1]} for col in content.dtypes]

    serialization_type = artifact_type_to_serialization_type(
        artifact_type, derived_from_bson, content
    ).value

    if output_path is not None:
        if artifact_type == ArtifactType.TABLE:
            # write artifact to temp view
            # take a sample of the DF
            # write that to aqueduct storage
            spark_df = content
            global_view_name = convert_path_to_view_name(output_path)
            spark_df.createOrReplaceGlobalTempView(global_view_name)
            pandas_df = spark_df.limit(100).toPandas()
            serialized_val = serialize_val_wrapper(pandas_df, serialization_type)
            storage.put(output_path, serialized_val)

        else:
            serialized_val = serialize_val_wrapper(content, serialization_type)
            storage.put(output_path, serialized_val)

    output_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = serialization_type
    output_metadata[_METADATA_PYTHON_TYPE_KEY] = type(content).__name__
    storage.put(output_metadata_path, json.dumps(output_metadata).encode(DEFAULT_ENCODING))


def convert_path_to_view_name(path: str) -> str:
    """
    Converts a given input/output path for an artifact and converts
    it into a Spark Temporary View compatible name. THis will convert
    slashes and hyphens into underscores.
    """
    return path.replace("-", "_").replace("/", "_")
