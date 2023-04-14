import io
import json
import os
import shutil
from typing import Any, Callable, Dict, List, Optional, Union, cast

import cloudpickle as pickle
import pandas as pd
from aqueduct.constants.enums import (
    ArtifactType,
    LocalDataSerializationType,
    LocalDataTableFormat,
    S3SerializationType,
    SerializationType,
)
from aqueduct.utils.local_data import _convert_to_local_data_table_format
from aqueduct.utils.type_inference import infer_artifact_type
from bson import json_util as bson_json_util
from PIL import Image
from pydantic import BaseModel

from .format import DEFAULT_ENCODING
from .function_packaging import _make_temp_dir

_DEFAULT_IMAGE_FORMAT = "jpeg"

# The temporary file name that a Tensorflow keras model will be dumped into before we read/write it from storage.
# This will be cleaned up within the serialization logic.
_TEMP_KERAS_MODEL_NAME = "keras_model"


class PickleableCollectionSerializationFormat(BaseModel):
    """For data types that are destined to be pickled lists or tuples, we want to
    first serialize each individual element before pickling, for performance reasons.

    When that happens, the dictionary version of this class is what is pickle-serialized.
    """

    # The serialization type of each element in the collection.
    aqueduct_serialization_types: Union[List[SerializationType], List[S3SerializationType]]

    # The actual list of serialized values.
    data: List[bytes]

    # Due to limitations of pydantic models with tuples, we need to have an explicit field
    # for this.
    is_tuple: bool


def _read_table_content(content: bytes) -> pd.DataFrame:
    return pd.read_json(io.BytesIO(content), orient="table")


def _read_bson_table_content(content: bytes) -> pd.DataFrame:
    return pd.DataFrame.from_records(bson_json_util.loads(content.decode(DEFAULT_ENCODING)))


def _read_json_content(content: bytes) -> Any:
    return json.loads(content.decode(DEFAULT_ENCODING))


def _read_pickle_content(content: bytes) -> Any:
    return pickle.loads(content)


def _read_image_content(content: bytes) -> Image.Image:
    return Image.open(io.BytesIO(content))


def _read_string_content(content: bytes) -> str:
    return content.decode(DEFAULT_ENCODING)


def _read_bytes_content(content: bytes) -> bytes:
    return content


def _read_local_csv_table_content(path: str) -> pd.DataFrame:
    return pd.read_csv(path)


def _read_local_json_table_content(path: str) -> pd.DataFrame:
    return pd.read_json(path, orient="table")


def _read_local_parquet_table_content(path: str) -> pd.DataFrame:
    return pd.read_parquet(path)


def _read_local_image_content(path: str) -> Image.Image:
    return Image.open(path)


def _read_local_json_content(path: str) -> Any:
    with open(path, mode="rb", encoding=DEFAULT_ENCODING) as file:
        return json.load(file)


def _read_local_pickle_content(path: str) -> Any:
    with open(path, mode="rb") as file:
        return pickle.load(file)


def _read_local_string_content(path: str) -> str:
    with open(path, mode="r", encoding=DEFAULT_ENCODING) as file:
        return file.read()


def _read_local_bytes_content(path: str) -> bytes:
    with open(path, mode="rb") as file:
        return file.read()


def _read_local_tf_keras_model(path: str) -> Any:
    from tensorflow import keras

    return keras.models.load_model(path)


# Returns a tf.keras.Model type. We don't assume that every user has it installed,
# so we return "Any" type.
def _read_tf_keras_model(content: bytes) -> Any:
    temp_model_dir = None
    try:
        temp_model_dir = _make_temp_dir()
        model_file_path = os.path.join(temp_model_dir, _TEMP_KERAS_MODEL_NAME)
        with open(model_file_path, "wb") as f:
            f.write(content)

        from tensorflow import keras

        return keras.load_model(model_file_path)
    finally:
        if temp_model_dir is not None and os.path.exists(temp_model_dir):
            shutil.rmtree(temp_model_dir)


# Not intended for use outside of `deserialize()`.
__deserialization_function_mapping: Dict[str, Callable[[bytes], Any]] = {
    SerializationType.TABLE: _read_table_content,
    SerializationType.JSON: _read_json_content,
    SerializationType.PICKLE: _read_pickle_content,
    SerializationType.IMAGE: _read_image_content,
    SerializationType.STRING: _read_string_content,
    SerializationType.BYTES: _read_bytes_content,
    SerializationType.TF_KERAS: _read_tf_keras_model,
    SerializationType.BSON_TABLE: _read_bson_table_content,
}

# Not intended for use outside of `deserialize()`.
__local_data_deserialization_function_mapping: Dict[str, Callable[[str], Any]] = {
    LocalDataSerializationType.CSV_TABLE: _read_local_csv_table_content,
    LocalDataSerializationType.JSON_TABLE: _read_local_json_table_content,
    LocalDataSerializationType.PARQUET_TABLE: _read_local_parquet_table_content,
    LocalDataSerializationType.IMAGE: _read_local_image_content,
    LocalDataSerializationType.JSON: _read_local_json_content,
    LocalDataSerializationType.BYTES: _read_local_bytes_content,
    LocalDataSerializationType.PICKLE: _read_local_pickle_content,
    LocalDataSerializationType.STRING: _read_local_string_content,
    LocalDataSerializationType.TF_KERAS: _read_local_tf_keras_model,
}


def check_and_fetch_pickled_collection_format(
    serialization_type: Union[SerializationType, S3SerializationType],
    deserialized_val: Any,
) -> Optional[PickleableCollectionSerializationFormat]:
    """If a value that has undergone one round of deserialization is in the form of a
    `PickleableCollectionSerializationFormat`, we will load up that class and return it.
    Otherwise, return None.
    """
    assert SerializationType.PICKLE.value == S3SerializationType.PICKLE.value
    if serialization_type == SerializationType.PICKLE and isinstance(deserialized_val, dict):
        try:
            # This will error if the appropriate dict fields do not match.
            return PickleableCollectionSerializationFormat(**deserialized_val)
        except Exception:
            return None
    return None


def deserialize(
    serialization_type: Union[SerializationType, S3SerializationType],
    artifact_type: ArtifactType,
    content: bytes,
    custom_deserialization_function_mapping: Optional[Dict[str, Callable[[bytes], Any]]] = None,
) -> Any:
    """Deserializes a byte string into the appropriate python object.

    Handles serialization for both the Aqueduct storage layer (default) and S3 (requires a custom deserialization function mapping).
    """
    deserialization_function_mapping = __deserialization_function_mapping
    if custom_deserialization_function_mapping is not None:
        assert isinstance(serialization_type, S3SerializationType)
        deserialization_function_mapping = custom_deserialization_function_mapping

    if serialization_type not in deserialization_function_mapping:
        raise Exception("Unsupported serialization type %s" % serialization_type)

    deserialized_val = deserialization_function_mapping[serialization_type](content)

    # Check if the type is a pickled collection where each individual element needs to be deserialized.
    pickled_collection_data = check_and_fetch_pickled_collection_format(
        serialization_type, deserialized_val
    )
    if pickled_collection_data is not None:
        collection_serialization_types = pickled_collection_data.aqueduct_serialization_types
        data = pickled_collection_data.data

        deserialized_val = [
            deserialization_function_mapping[collection_serialization_types[i]](data[i])
            for i in range(len(collection_serialization_types))
        ]

        if pickled_collection_data.is_tuple:
            return tuple(deserialized_val)
        return deserialized_val

    # Because both list and tuple objects are json-serialized, they will have the same bytes representation.
    # We wanted to keep the readability of json, particularly for the UI, so we decided to distinguish
    # between the two here using the expected artifact type, at deserialization time.
    if artifact_type == ArtifactType.TUPLE:
        return tuple(deserialized_val)
    return deserialized_val


def deserialize_from_local_data(
    serialization_type: LocalDataSerializationType,
    artifact_type: ArtifactType,
    path: str,
) -> Any:
    """Deserializes a file object with specified path into the appropriate python object.

    Handles serialization for local data.
    """
    deserialization_function_mapping = __local_data_deserialization_function_mapping
    if serialization_type not in deserialization_function_mapping:
        raise Exception("Unsupported serialization type %s" % serialization_type)

    deserialized_val = deserialization_function_mapping[serialization_type](path)

    if artifact_type == ArtifactType.TUPLE:
        return tuple(deserialized_val)
    return deserialized_val


def _write_table_output(output: pd.DataFrame) -> bytes:
    # This serialization format should also be consistent with go code in
    # src/golang/lib/workflow/artifact/artifact.go SampleContent() method.
    output_str = cast(str, output.to_json(orient="table", date_format="iso", index=False))
    return output_str.encode(DEFAULT_ENCODING)


def _write_bson_table_output(output: pd.DataFrame) -> bytes:
    # This serialization format should also be consistent with go code in
    # src/golang/lib/workflow/artifact/artifact.go SampleContent() method.
    return bson_json_util.dumps(output.to_dict(orient="records")).encode(DEFAULT_ENCODING)


def _write_image_output(output: Image.Image) -> bytes:
    img_bytes = io.BytesIO()
    output.save(img_bytes, format=_DEFAULT_IMAGE_FORMAT)
    return img_bytes.getvalue()


def _write_string_output(output: str) -> bytes:
    return output.encode(DEFAULT_ENCODING)


def _write_bytes_output(output: bytes) -> bytes:
    return output


def _write_pickle_output(output: Any) -> bytes:
    return bytes(pickle.dumps(output))


def _write_json_output(output: Any) -> bytes:
    return json.dumps(output).encode(DEFAULT_ENCODING)


def _write_tf_keras_model(output: Any) -> bytes:
    temp_model_dir = None
    try:
        temp_model_dir = _make_temp_dir()
        model_file_path = os.path.join(temp_model_dir, _TEMP_KERAS_MODEL_NAME)

        output.save(model_file_path)
        return open(model_file_path, "rb").read()
    finally:
        if temp_model_dir is not None and os.path.exists(temp_model_dir):
            shutil.rmtree(temp_model_dir)


__serialization_function_mapping: Dict[str, Callable[..., bytes]] = {
    SerializationType.TABLE: _write_table_output,
    SerializationType.JSON: _write_json_output,
    SerializationType.PICKLE: _write_pickle_output,
    SerializationType.IMAGE: _write_image_output,
    SerializationType.STRING: _write_string_output,
    SerializationType.BYTES: _write_bytes_output,
    SerializationType.TF_KERAS: _write_tf_keras_model,
    SerializationType.BSON_TABLE: _write_bson_table_output,
}


def serialize_val(
    val: Any,
    serialization_type: Union[SerializationType, S3SerializationType],
    derived_from_bson: bool,
) -> bytes:
    """Serializes a parameter or computed value into bytes."""
    if serialization_type == SerializationType.PICKLE and (
        isinstance(val, list) or isinstance(val, tuple)
    ):
        elem_serialization_types: List[SerializationType] = []
        for elem in val:
            elem_artifact_type = infer_artifact_type(elem)
            elem_serialization_types.append(
                artifact_type_to_serialization_type(
                    elem_artifact_type,
                    derived_from_bson,
                    elem,
                ),
            )

        data: List[bytes] = [
            __serialization_function_mapping[elem_serialization_types[i]](val[i])
            for i in range(len(elem_serialization_types))
        ]

        pickled_collection_data = PickleableCollectionSerializationFormat(
            aqueduct_serialization_types=elem_serialization_types,
            data=data,
            is_tuple=isinstance(val, tuple),
        )

        # The value we end up pickling is a dictionary.
        val = pickled_collection_data.dict()

    return __serialization_function_mapping[serialization_type](val)


def artifact_type_to_serialization_type(
    artifact_type: ArtifactType,
    # derived_from_bson specifies if the artifact is derived from a bson object
    # and thus requires bson encoding.
    # For now, it only applies to DataFrames extracted / transformed from Mongo.
    derived_from_bson: bool,
    content: Any,
) -> SerializationType:
    if artifact_type == ArtifactType.TABLE:
        serialization_type = (
            SerializationType.BSON_TABLE if derived_from_bson else SerializationType.TABLE
        )
    elif artifact_type == ArtifactType.IMAGE:
        serialization_type = SerializationType.IMAGE
    elif artifact_type == ArtifactType.JSON or artifact_type == ArtifactType.STRING:
        serialization_type = SerializationType.STRING
    elif artifact_type == ArtifactType.BYTES:
        serialization_type = SerializationType.BYTES
    elif artifact_type == ArtifactType.BOOL or artifact_type == ArtifactType.NUMERIC:
        serialization_type = SerializationType.JSON
    elif artifact_type == ArtifactType.PICKLABLE:
        serialization_type = SerializationType.PICKLE
    elif (
        artifact_type == ArtifactType.DICT
        or artifact_type == ArtifactType.TUPLE
        or artifact_type == ArtifactType.LIST
    ):
        try:
            json.dumps(content)
            serialization_type = SerializationType.JSON
        except:
            serialization_type = SerializationType.PICKLE
    elif artifact_type == ArtifactType.TF_KERAS:
        serialization_type = SerializationType.TF_KERAS
    else:
        raise Exception("Unsupported artifact type %s" % artifact_type)

    assert serialization_type is not None, (
        "Unimplemented case for artifact type `%s`" % artifact_type
    )
    return serialization_type


def extract_val_from_local_data(
    path: str, as_type: Optional[ArtifactType], format: Optional[str]
) -> Any:
    """Extract value of specified type in Local Data."""
    artifact_type = as_type
    local_data_path = path
    local_data_format = _convert_to_local_data_table_format(format)
    if artifact_type == ArtifactType.TABLE:
        if local_data_format == LocalDataTableFormat.CSV:
            local_data_serialization_format = LocalDataSerializationType.CSV_TABLE
        elif local_data_format == LocalDataTableFormat.JSON:
            local_data_serialization_format = LocalDataSerializationType.JSON_TABLE
        elif local_data_format == LocalDataTableFormat.PARQUET:
            local_data_serialization_format = LocalDataSerializationType.PARQUET_TABLE
        else:
            raise Exception("Unsupported file format %s" % format)
    elif artifact_type == ArtifactType.IMAGE:
        local_data_serialization_format = LocalDataSerializationType.IMAGE
    elif artifact_type == ArtifactType.JSON or artifact_type == ArtifactType.STRING:
        local_data_serialization_format = LocalDataSerializationType.STRING
    elif artifact_type == ArtifactType.BYTES:
        local_data_serialization_format = LocalDataSerializationType.BYTES
    elif artifact_type == ArtifactType.BOOL or artifact_type == ArtifactType.NUMERIC:
        local_data_serialization_format = LocalDataSerializationType.JSON
    elif artifact_type == ArtifactType.PICKLABLE:
        local_data_serialization_format = LocalDataSerializationType.PICKLE
    elif (
        artifact_type == ArtifactType.DICT
        or artifact_type == ArtifactType.TUPLE
        or artifact_type == ArtifactType.LIST
    ):
        try:
            with open(local_data_path, mode="rb") as file:
                json.dumps(file.read())
            local_data_serialization_format = LocalDataSerializationType.JSON
        except:
            local_data_serialization_format = LocalDataSerializationType.PICKLE
    elif artifact_type == ArtifactType.TF_KERAS:
        local_data_serialization_format = LocalDataSerializationType.TF_KERAS
    else:
        raise Exception("Unsupported artifact type %s" % artifact_type)

    deserialized_val = deserialize_from_local_data(
        local_data_serialization_format, artifact_type, local_data_path
    )
    return deserialized_val
