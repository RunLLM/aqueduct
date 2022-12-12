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
from aqueduct_executor.operators.utils.enums import ArtifactType, SerializationType
from PIL import Image

_DEFAULT_ENCODING = "utf8"

# The temporary file name that a Tensorflow keras model will be dumped into before we read/write it from storage.
# This will be cleaned up within the serialization logic.
_TEMP_KERAS_MODEL_NAME = "keras_model"


def _read_table_content(content: bytes) -> pd.DataFrame:
    return pd.read_json(io.BytesIO(content), orient="table")


def _read_json_content(content: bytes) -> Any:
    return json.loads(content.decode(_DEFAULT_ENCODING))


def _read_pickle_content(content: bytes) -> Any:
    return pickle.loads(content)


def _read_image_content(content: bytes) -> Image.Image:
    return Image.open(io.BytesIO(content))


def _read_string_content(content: bytes) -> str:
    return content.decode(_DEFAULT_ENCODING)


def _read_bytes_content(content: bytes) -> bytes:
    return content


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


deserialization_function_mapping: Dict[str, Callable[[bytes], Any]] = {
    SerializationType.TABLE: _read_table_content,
    SerializationType.JSON: _read_json_content,
    SerializationType.PICKLE: _read_pickle_content,
    SerializationType.IMAGE: _read_image_content,
    SerializationType.STRING: _read_string_content,
    SerializationType.BYTES: _read_bytes_content,
    SerializationType.TF_KERAS: _read_tf_keras_model,
}
