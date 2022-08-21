import io
import json
import base64
from typing import Any

import cloudpickle as pickle
import pandas as pd
from aqueduct.enums import SerializationType
from PIL import Image

_DEFAULT_ENCODING_JSON = "utf8"
_DEFAULT_ENCODING_IMAGE = "JPEG"


def _write_table_content(content: pd.DataFrame) -> bytes:
    json_table = content.to_json(orient="table", date_format="iso", index=False)
    return json.dumps(json_table).encode(_DEFAULT_ENCODING_JSON)


def _write_json_content(content: Any) -> bytes:
    return json.dumps(content).encode(_DEFAULT_ENCODING_JSON)


def _write_pickle_content(content: bytes) -> Any:
    return pickle.loads(content)


def _write_image_content(content: Image.Image) -> bytes:
    image_buffer = io.BytesIO()
    content.save(image_buffer, format=_DEFAULT_ENCODING_IMAGE)
    return image_buffer.getvalue()


def _write_string_content(content: str) -> bytes:
    return content.encode(_DEFAULT_ENCODING_JSON)


def _write_bytes_content(content: bytes) -> bytes:
    return content


serialization_function_mapping = {
    SerializationType.TABLE: _write_table_content,
    SerializationType.JSON: _write_json_content,
    SerializationType.PICKLE: _write_pickle_content,
    SerializationType.IMAGE: _write_image_content,
    SerializationType.STRING: _write_string_content,
    SerializationType.BYTES: _write_bytes_content,
}