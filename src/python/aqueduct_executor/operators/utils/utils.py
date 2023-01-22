import io
import json
from typing import Any, Callable, Dict, List, Optional, Tuple
from pyspark.sql import SparkSession

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

_RUNTIME_SEC_METRIC_NAME = "runtime"
_MAX_MEMORY_MB_METRIC_NAME = "max_memory"
_METADATA_SCHEMA_KEY = "schema"
_METADATA_SYSTEM_METADATA_KEY = "system_metadata"
_METADATA_ARTIFACT_TYPE_KEY = "artifact_type"
_METADATA_SERIALIZATION_TYPE_KEY = "serialization_type"
_METADATA_PYTHON_TYPE_KEY = "python_type"

# The temporary file name that a Tensorflow keras model will be dumped into before we read/write it from storage.
# This will be cleaned up within the serialization logic.
_TEMP_KERAS_MODEL_NAME = "keras_model"


def _read_csv(input_bytes: bytes) -> pd.DataFrame:
    return pd.read_csv(io.BytesIO(input_bytes))


def _read_json_bytes(input_bytes: bytes) -> Any:
    return json.loads(input_bytes.decode(DEFAULT_ENCODING))


def read_artifacts(
    storage: Storage,
    input_paths: List[str],
    input_metadata_paths: List[str],
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
        # Make sure that the input paths exist.
        try:
            _ = storage.get(input_path)
            _ = storage.get(input_metadata_path)
        except Exception as e:
            # TODO(ENG-1627): think about retrying the parent operator in such instances.
            raise MissingInputPathsException(
                "Unable to read inputs artifacts. Exception: %s" % str(e)
            )

        artifact_metadata = json.loads(storage.get(input_metadata_path).decode(DEFAULT_ENCODING))
        artifact_type = artifact_metadata[_METADATA_ARTIFACT_TYPE_KEY]
        artifact_types.append(artifact_type)

        serialization_type = artifact_metadata[_METADATA_SERIALIZATION_TYPE_KEY]
        serialization_types.append(serialization_type)

        inputs.append(deserialize(serialization_type, artifact_type, storage.get(input_path)))

    return inputs, artifact_types, serialization_types


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
            #read from temp view
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


def read_system_metadata(
    storage: Storage,
    input_metadata_paths: List[str],
) -> List[Dict[str, Any]]:
    return _read_metadata_key(storage, input_metadata_paths, _METADATA_SYSTEM_METADATA_KEY)


def _read_metadata_key(
    storage: Storage, input_metadata_paths: List[str], key_name: str
) -> List[Dict[str, Any]]:
    metadata_inputs = [
        _read_json_bytes(storage.get(input_path)) for input_path in input_metadata_paths
    ]
    if any(key_name not in metadata for metadata in metadata_inputs):
        raise Exception(key_name + " does not exist in input metadata.")
    return [metadata[key_name] for metadata in metadata_inputs]


def serialize_val_wrapper(val: Any, serialization_type: SerializationType) -> bytes:
    """Wrapper around `serialize_val()` to perform additional checks that are specific to the executor."""
    if serialization_type == SerializationType.TABLE:
        # We cannot serialize integer column names into json.
        violating_col_names = [col for col in val.columns if not isinstance(col, str)]
        if len(violating_col_names) > 0:
            raise ExecFailureException(
                failure_type=FailureType.USER_FATAL,
                tip="Non-String column names are not supported. Violating columns: %s"
                % (", ".join(violating_col_names)),
            )

    serialized_val = serialize_val(val, serialization_type)
    assert isinstance(serialized_val, bytes)  # Necessary for mypy
    return serialized_val


def write_artifact(
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
) -> None:
    """The `output_path` can be empty if the contents were already pre-populated (eg. parameter operators)."""
    output_metadata: Dict[str, Any] = {
        _METADATA_SCHEMA_KEY: [],
        _METADATA_SYSTEM_METADATA_KEY: system_metadata,
        _METADATA_ARTIFACT_TYPE_KEY: artifact_type.value,
    }

    if artifact_type == ArtifactType.TABLE:
        output_metadata[_METADATA_SCHEMA_KEY] = [{col: str(content[col].dtype)} for col in content]

    serialization_type = artifact_type_to_serialization_type(
        artifact_type, derived_from_bson, content
    ).value

    if output_path is not None:
        serialized_val = serialize_val_wrapper(content, serialization_type)
        storage.put(output_path, serialized_val)

    output_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = serialization_type
    output_metadata[_METADATA_PYTHON_TYPE_KEY] = type(content).__name__
    storage.put(output_metadata_path, json.dumps(output_metadata).encode(DEFAULT_ENCODING))


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
            print("global_view_name")
            print(global_view_name)
            spark_df.createOrReplaceGlobalTempView(global_view_name)
            # sampleDF = sparkDF.sample(fraction=0.1)
            pandas_df = spark_df.limit(100).toPandas()
            serialized_val = serialize_val_wrapper(pandas_df, serialization_type)
            storage.put(output_path, serialized_val)
            
        else:
            serialized_val = serialize_val_wrapper(content, serialization_type)
            storage.put(output_path, serialized_val)

    output_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = serialization_type
    output_metadata[_METADATA_PYTHON_TYPE_KEY] = type(content).__name__
    storage.put(output_metadata_path, json.dumps(output_metadata).encode(DEFAULT_ENCODING))


def write_exec_state(
    storage: Storage,
    metadata_path: str,
    exec_state: ExecutionState,
) -> None:
    """
    Writes operator execution logs to storage.
    :param err: Any error message encountered during execution.
    :param logs: Any logs generated by this operator.
    """
    storage.put(metadata_path, bytes(exec_state.json(), encoding=DEFAULT_ENCODING))


def delete_object(name: str, delete_fn: Callable[[str], None]) -> SavedObjectDelete:
    exec_state = ExecutionState(user_logs=Logs())
    try:
        delete_fn(name)
    except Exception as e:
        exec_state.status = ExecutionStatus.FAILED
        exec_state.failure_type = FailureType.SYSTEM
        exec_state.error = Error(context=exception_traceback(e), tip=TIP_UNKNOWN_ERROR)
        return SavedObjectDelete(name=name, exec_state=exec_state)
    exec_state.status = ExecutionStatus.SUCCEEDED
    return SavedObjectDelete(name=name, exec_state=exec_state)


def write_delete_saved_objects_results(
    storage: Storage, path: str, results: Dict[str, List[SavedObjectDelete]]
) -> None:
    # `[object].json()`` gives me the json as a string which gets escaped in `json.dumps()``.
    # However, I cannot serialize it directly with `json.dumps()`` so I serialize with
    # `[object].json()` then `json.loads()` to convert it to a dictionary that is serializable
    # by `json.dumps.()`.
    results_str = json.dumps(
        {
            integration: [
                {"name": result.name, "exec_state": json.loads(result.exec_state.json())}
                for result in results[integration]
            ]
            for integration in results
        }
    )
    storage.put(path, bytes(results_str, encoding=DEFAULT_ENCODING))


def write_discover_results(storage: Storage, path: str, tables: List[str]) -> None:
    table_names_str = json.dumps(tables)

    storage.put(path, bytes(table_names_str, encoding=DEFAULT_ENCODING))


def write_compile_airflow_output(storage: Storage, path: str, dag_file: bytes) -> None:
    """
    Writes the provided Airflow DAG file to storage.
    """
    storage.put(path, dag_file)


def convert_path_to_view_name(path: str) -> str:
    """
    Converts a given input/output path for an artifact and converts
    it into a Spark Temporary View compatible name. THis will convert
    slashes and hyphens into underscores.
    """
    return path.replace("-", "_").replace("/", "_")
