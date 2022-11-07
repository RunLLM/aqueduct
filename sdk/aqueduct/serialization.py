import base64
import io
import json
import os
import shutil
import tempfile
import uuid
from pathlib import Path
from typing import Any, Callable, Dict, cast

import cloudpickle as pickle
import pandas as pd
from aqueduct.enums import ArtifactType, SerializationType
from PIL import Image

DEFAULT_ENCODING = "utf8"
_DEFAULT_IMAGE_FORMAT = "jpeg"

# The temporary file name that a Tensorflow keras model will be dumped into before we read/write it from storage.
# This will be cleaned up within the serialization logic.
_TEMP_KERAS_MODEL_NAME = "keras_model"


def _read_table_content(content: bytes) -> pd.DataFrame:
    return pd.read_json(io.BytesIO(content), orient="table")


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


# NOTE: this doesn't really belong in the serialization library, but we don't have a lower layer than this right now -
# (utils.py imports this file).
def make_temp_dir() -> str:
    """
    Create a unique, temporary directory in the local filesystem and returns the path.
    """
    dir_path = None
    created = False
    # Try to create the directory. If it already exists, try again with a new name.
    while not created:
        dir_path = Path(tempfile.gettempdir()) / str(uuid.uuid4())
        try:
            os.mkdir(dir_path)
            created = True
        except FileExistsError:
            pass

    assert dir_path is not None
    return str(dir_path)


# Returns a tf.keras.Model type. We don't assume that every user has it installed,
# so we return "Any" type.
def _read_tf_keras_model(content: bytes) -> Any:
    temp_model_dir = None
    try:
        temp_model_dir = make_temp_dir()
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
}


def deserialize(
    serialization_type: SerializationType, artifact_type: ArtifactType, content: bytes
) -> Any:
    """Deserializes a byte string into the appropriate python object."""
    if serialization_type not in __deserialization_function_mapping:
        raise Exception("Unsupported serialization type %s" % serialization_type)

    deserialized_val = __deserialization_function_mapping[serialization_type](content)

    # Because both list and tuple objects are json-serialized, they will have the same bytes representation.
    # We wanted to keep the readability of json, particularly for the UI, so we decided to distinguish
    # between the two here using the expected artifact type, at deserialization time.
    if artifact_type == ArtifactType.TUPLE:
        return tuple(deserialized_val)
    return deserialized_val


def _write_table_output(output: pd.DataFrame) -> bytes:
    output_str = cast(str, output.to_json(orient="table", date_format="iso", index=False))
    return output_str.encode(DEFAULT_ENCODING)


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
        temp_model_dir = make_temp_dir()
        model_file_path = os.path.join(temp_model_dir, _TEMP_KERAS_MODEL_NAME)

        output.save(model_file_path)
        return open(model_file_path, "rb").read()
    finally:
        if temp_model_dir is not None and os.path.exists(temp_model_dir):
            shutil.rmtree(temp_model_dir)


serialization_function_mapping: Dict[str, Callable[..., bytes]] = {
    SerializationType.TABLE: _write_table_output,
    SerializationType.JSON: _write_json_output,
    SerializationType.PICKLE: _write_pickle_output,
    SerializationType.IMAGE: _write_image_output,
    SerializationType.STRING: _write_string_output,
    SerializationType.BYTES: _write_bytes_output,
    SerializationType.TF_KERAS: _write_tf_keras_model,
}


def serialize_val(val: Any, serialization_type: SerializationType) -> bytes:
    """Serializes a parameter or computed value into bytes."""
    return serialization_function_mapping[serialization_type](val)


def artifact_type_to_serialization_type(
    artifact_type: ArtifactType, content: Any
) -> SerializationType:
    if artifact_type == ArtifactType.TABLE:
        serialization_type = SerializationType.TABLE
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
