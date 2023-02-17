import io
import json
import os
import shutil
from typing import Any, Callable, Dict, List, Optional, Tuple, Union, cast

import cloudpickle as pickle
import pandas as pd
from aqueduct.constants.enums import ArtifactType, SerializationType
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
    aqueduct_serialization_types: List[SerializationType]

    # The actual list of serialized values.
    data: Union[List[bytes], Tuple[bytes]]


def _serialization_is_pickle(serialization_type: SerializationType) -> bool:
    return (
        serialization_type == SerializationType.PICKLE
        or serialization_type == SerializationType.BSON_PICKLE
    )


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
    SerializationType.BSON_PICKLE: _read_pickle_content,
    SerializationType.IMAGE: _read_image_content,
    SerializationType.STRING: _read_string_content,
    SerializationType.BYTES: _read_bytes_content,
    SerializationType.TF_KERAS: _read_tf_keras_model,
    SerializationType.BSON_TABLE: _read_bson_table_content,
}


def check_and_fetch_pickled_collection_format(
    serialization_type: SerializationType,
    deserialized_val: Any,
) -> Optional[PickleableCollectionSerializationFormat]:
    """If a value that has undergone one round of deserialization is in the form of a
    `PickleableCollectionSerializationFormat`, we will load up that class and return it.
    Otherwise, return None.
    """
    if _serialization_is_pickle(serialization_type) and isinstance(deserialized_val, dict):
        try:
            # This will error if the appropriate dict fields do not match.
            return PickleableCollectionSerializationFormat(**deserialized_val)
        except Exception:
            return None
    return None


def deserialize(
    serialization_type: SerializationType, artifact_type: ArtifactType, content: bytes
) -> Any:
    """Deserializes a byte string into the appropriate python object."""
    if serialization_type not in __deserialization_function_mapping:
        raise Exception("Unsupported serialization type %s" % serialization_type)

    deserialized_val = __deserialization_function_mapping[serialization_type](content)

    # Check if the type is an expanded collection and resolve the content for that special case.
    pickled_collection_data = check_and_fetch_pickled_collection_format(
        serialization_type, deserialized_val
    )
    if pickled_collection_data is not None:
        collection_serialization_types = pickled_collection_data.aqueduct_serialization_types
        data = pickled_collection_data.data

        deserialized_val = [
            __deserialization_function_mapping[collection_serialization_types[i]](data[i])
            for i in range(len(collection_serialization_types))
        ]

        if isinstance(data, tuple):
            return tuple(deserialized_val)
        return deserialized_val

    # Because both list and tuple objects are json-serialized, they will have the same bytes representation.
    # We wanted to keep the readability of json, particularly for the UI, so we decided to distinguish
    # between the two here using the expected artifact type, at deserialization time.
    if artifact_type == ArtifactType.TUPLE:
        return tuple(deserialized_val)
    return deserialized_val


def _write_table_output(output: pd.DataFrame) -> bytes:
    output_str = cast(str, output.to_json(orient="table", date_format="iso", index=False))
    return output_str.encode(DEFAULT_ENCODING)


def _write_bson_table_output(output: pd.DataFrame) -> bytes:
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


serialization_function_mapping: Dict[str, Callable[..., bytes]] = {
    SerializationType.TABLE: _write_table_output,
    SerializationType.JSON: _write_json_output,
    SerializationType.PICKLE: _write_pickle_output,
    SerializationType.BSON_PICKLE: _write_pickle_output,
    SerializationType.IMAGE: _write_image_output,
    SerializationType.STRING: _write_string_output,
    SerializationType.BYTES: _write_bytes_output,
    SerializationType.TF_KERAS: _write_tf_keras_model,
    SerializationType.BSON_TABLE: _write_bson_table_output,
}


def serialize_val(
    val: Any, serialization_type: SerializationType, expand_collections: bool = True
) -> bytes:
    """Serializes a parameter or computed value into bytes."""
    if (
        expand_collections
        and _serialization_is_pickle(serialization_type)
        and (isinstance(val, list) or isinstance(val, tuple))
    ):
        elem_serialization_types: List[SerializationType] = []
        for elem in val:
            elem_artifact_type = infer_artifact_type(elem)
            elem_serialization_types.append(
                artifact_type_to_serialization_type(
                    elem_artifact_type,
                    serialization_type == SerializationType.BSON_PICKLE,  # derived_from_bson
                    elem,
                )
            )
        data: Union[List[bytes], Tuple[bytes]] = [
            serialize_val(
                val[i], elem_serialization_types[i], expand_collections=False
            )  # do not recursively expand.
            for i in range(len(elem_serialization_types))
        ]
        if isinstance(val, tuple):
            data = cast(Tuple[bytes], tuple(data))

        pickled_collection_data = PickleableCollectionSerializationFormat(
            data=data,
            aqueduct_serialization_types=elem_serialization_types,
        )

        # The value we end up serializing is a dictionary.
        val = pickled_collection_data.dict()

    return serialization_function_mapping[serialization_type](val)


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
        serialization_type = (
            SerializationType.BSON_PICKLE if derived_from_bson else SerializationType.PICKLE
        )
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
