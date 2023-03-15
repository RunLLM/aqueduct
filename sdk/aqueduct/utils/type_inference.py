import base64
import json
from typing import Any

import cloudpickle as pickle
import numpy as np
from aqueduct.constants.enums import ArtifactType
from pandas import DataFrame
from PIL import Image

from .format import DEFAULT_ENCODING


def infer_artifact_type(value: Any) -> ArtifactType:
    if isinstance(value, DataFrame):
        return ArtifactType.TABLE
    elif isinstance(value, Image.Image):
        return ArtifactType.IMAGE
    elif isinstance(value, bytes):
        return ArtifactType.BYTES
    elif isinstance(value, str):
        # We first check if the value is a valid JSON string.
        try:
            json.loads(value)
            return ArtifactType.JSON
        except:
            return ArtifactType.STRING
    elif isinstance(value, bool) or isinstance(value, np.bool_):
        return ArtifactType.BOOL
    elif isinstance(value, int) or isinstance(value, float) or isinstance(value, np.number):
        return ArtifactType.NUMERIC
    elif isinstance(value, dict):
        return ArtifactType.DICT
    elif isinstance(value, tuple):
        return ArtifactType.TUPLE
    elif isinstance(value, list):
        return ArtifactType.LIST
    else:
        try:
            pickle.dumps(value)
            return ArtifactType.PICKLABLE
        except:
            pass

        try:
            # tf.keras.Model's can be pickled, but some classes that inherit from it cannot (eg. `tfrs.Model`)
            from tensorflow import keras

            if isinstance(value, keras.Model):
                return ArtifactType.TF_KERAS
        except:
            pass

        raise Exception("Failed to map type %s to supported artifact type." % type(value))


def _bytes_to_base64_string(content: bytes) -> str:
    """Helper to convert bytes to a base64-string.

    For example, image-serialized bytes are not `utf8` encoded, so if we want to convert
    such bytes to string, we must use this function.
    """
    return base64.b64encode(content).decode(DEFAULT_ENCODING)


def _base64_string_to_bytes(content: str) -> bytes:
    """Helpers to convert base64-string back to bytes."""
    return base64.b64decode(content)
