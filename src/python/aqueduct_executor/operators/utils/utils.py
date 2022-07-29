import io
import json
from typing import Any, Dict, List, Union, Tuple

import numpy as np
import pandas as pd
import cloudpickle as pickle
from PIL import Image
from aqueduct_executor.operators.utils.enums import ArtifactType, SerializationType, artifact_to_serialization
from aqueduct_executor.operators.utils.execution import ExecutionState
from aqueduct_executor.operators.utils.storage.storage import Storage

_DEFAULT_ENCODING = "utf8"
_DEFAULT_IMAGE_FORMAT = "jpeg"
_RUNTIME_SEC_METRIC_NAME = "runtime"
_MAX_MEMORY_MB_METRIC_NAME = "max_memory"
_METADATA_SCHEMA_KEY = "schema"
_METADATA_SYSTEM_METADATA_KEY = "system_metadata"
_METADATA_ARTIFACT_TYPE_KEY = "artifact_type"
_METADATA_SERIALIZATION_TYPE_KEY = "serialization_type"


def _read_csv(storage: Storage, path: str) -> pd.DataFrame:
    input_bytes = storage.get(path)
    return pd.read_csv(io.BytesIO(input_bytes))


def _read_tabular_input(storage: Storage, path: str) -> pd.DataFrame:
    input_bytes = storage.get(path)
    return pd.read_json(io.BytesIO(input_bytes), orient="table")


def _read_json_input(storage: Storage, path: str) -> Any:
    return json.loads(storage.get(path).decode(_DEFAULT_ENCODING))


def _read_pickle_input(storage: Storage, path: str) -> Any:
    return pickle.loads(storage.get(path))


def _read_image_input(storage: Storage, path: str) -> Image.Image:
    return Image.open(io.BytesIO(storage.get(path)))


def _read_string_input(storage: Storage, path: str) -> str:
    return storage.get(path).decode(_DEFAULT_ENCODING)


def _read_bytes_input(storage: Storage, path: str) -> bytes:
    return storage.get(path)


_deserialization_function_mapping = {
    SerializationType.TABULAR: _read_tabular_input,
    SerializationType.JSON: _read_json_input,
    SerializationType.PICKLE: _read_pickle_input,
    SerializationType.IMAGE: _read_image_input,
    SerializationType.STRING: _read_string_input,
    SerializationType.BYTES: _read_bytes_input,
}


def read_artifacts(
    storage: Storage,
    input_paths: List[str],
    input_metadata_paths: List[str],
) -> Tuple[List[Any], List[ArtifactType]]:
    if len(input_paths) != len(input_metadata_paths):
        raise Exception(
            "Found inconsistent number of input paths (%d) and input metadata paths (%d)"
            % (
                len(input_paths),
                len(input_metadata_paths),
            )
        )

    inputs: List[Any] = []
    input_types: List[ArtifactType] = []

    for (input_path, input_metadata_path) in zip(input_paths, input_metadata_paths):
        artifact_metadata = json.loads(storage.get(input_metadata_path).decode(_DEFAULT_ENCODING))
        artifact_type = artifact_metadata[_METADATA_ARTIFACT_TYPE_KEY]
        input_types.append(artifact_type)

        serialization_type = artifact_metadata[_METADATA_SERIALIZATION_TYPE_KEY]
        if serialization_type not in _deserialization_function_mapping:
            raise Exception("Unsupported serialization type %s" % serialization_type)
        inputs.append(_deserialization_function_mapping[serialization_type](storage, input_path))

    return inputs, input_types


def read_system_metadata(
    storage: Storage,
    input_metadata_paths: List[str],
) -> List[Dict[str, Any]]:
    return _read_metadata_key(storage, input_metadata_paths, _METADATA_SYSTEM_METADATA_KEY)


def _read_metadata_key(
    storage: Storage, input_metadata_paths: List[str], key_name: str
) -> List[Dict[str, Any]]:
    metadata_inputs = [_read_json_input(storage, input_path) for input_path in input_metadata_paths]
    if any(key_name not in metadata for metadata in metadata_inputs):
        raise Exception(key_name + " does not exist in input metadata.")
    return [metadata[key_name] for metadata in metadata_inputs]


def _write_tabular_output(
    storage: Storage,
    output_path: str,
    output: pd.DataFrame,
) -> None:
    output_str = output.to_json(orient="table", date_format="iso", index=False)
    storage.put(output_path, output_str.encode(_DEFAULT_ENCODING))


def _write_image_output(
    storage: Storage,
    output_path: str,
    output: Image.Image,
) -> None:
    img_bytes = io.BytesIO()
    output.save(img_bytes, format=_DEFAULT_IMAGE_FORMAT)
    storage.put(output_path, img_bytes.getvalue())


def _write_string_output(
    storage: Storage,
    output_path: str,
    output: str,
) -> None:
    storage.put(output_path, output.encode(_DEFAULT_ENCODING))


def _write_bytes_output(
    storage: Storage,
    output_path: str,
    output: bytes,
) -> None:
    storage.put(output_path, output)


def _write_pickle_output(
    storage: Storage,
    output_path: str,
    output: Any,
) -> None:
    storage.put(output_path, pickle.dumps(output))


def _write_json_output(
    storage: Storage,
    output_path: str,
    output: Any,
) -> None:
    storage.put(output_path, json.dumps(output).encode(_DEFAULT_ENCODING))


_serialization_function_mapping = {
    SerializationType.TABULAR: _write_tabular_output,
    SerializationType.JSON: _write_json_output,
    SerializationType.PICKLE: _write_pickle_output,
    SerializationType.IMAGE: _write_image_output,
    SerializationType.STRING: _write_string_output,
    SerializationType.BYTES: _write_bytes_output,
}


def write_artifact(
    storage: Storage,
    artifact_type: ArtifactType,
    output_path: str,
    output_metadata_path: str,
    content: Any,
    system_metadata: Dict[str, str],
) -> None:
    output_metadata: Dict[str, Any] = {
        _METADATA_SCHEMA_KEY: [],
        _METADATA_SYSTEM_METADATA_KEY: system_metadata,
        _METADATA_ARTIFACT_TYPE_KEY: artifact_type.value,
    }

    if artifact_type == ArtifactType.TABULAR:
        output_metadata[_METADATA_SCHEMA_KEY] = [{col: str(content[col].dtype)} for col in content]
        output_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = SerializationType.TABULAR.value
    elif artifact_type == ArtifactType.IMAGE:
        output_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = SerializationType.IMAGE.value
    elif artifact_type == ArtifactType.JSON or artifact_type == ArtifactType.STRING:
        output_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = SerializationType.STRING.value
    elif artifact_type == ArtifactType.BYTES:
        output_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = SerializationType.BYTES.value
    elif artifact_type == ArtifactType.BOOL or artifact_type == ArtifactType.NUMERIC:
        output_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = SerializationType.JSON.value
    elif artifact_type == ArtifactType.PICKLABLE:
        output_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = SerializationType.PICKLE.value
    elif artifact_type == ArtifactType.DICT or artifact_type == ArtifactType.TUPLE:
        try:
            json.dumps(content)
            output_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = SerializationType.JSON.value
        except:
            output_metadata[_METADATA_SERIALIZATION_TYPE_KEY] = SerializationType.PICKLE.value
    else:
        raise Exception("Unsupported artifact type %s" % artifact_type)

    assert(output_metadata[_METADATA_SERIALIZATION_TYPE_KEY] in artifact_to_serialization[artifact_type])
    _serialization_function_mapping[output_metadata[_METADATA_SERIALIZATION_TYPE_KEY]](
        storage,
        output_path,
        content,
    )
    
    storage.put(
        output_metadata_path, json.dumps(output_metadata).encode(_DEFAULT_ENCODING)
    )


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
    storage.put(metadata_path, bytes(exec_state.json(), encoding=_DEFAULT_ENCODING))


def write_discover_results(storage: Storage, path: str, tables: List[str]) -> None:
    table_names_str = json.dumps(tables)

    storage.put(path, bytes(table_names_str, encoding=_DEFAULT_ENCODING))
