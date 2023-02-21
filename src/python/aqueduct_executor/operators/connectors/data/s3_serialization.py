import io
import json
import pickle
from typing import Any, Callable, Dict

import pandas as pd
from aqueduct.constants.enums import S3SerializationType
from aqueduct_executor.operators.connectors.data import common, extract
from PIL import Image

from ...utils.enums import ArtifactType
from .s3 import _DEFAULT_JSON_ENCODING


def artifact_type_to_s3_serialization_type(
    key: str,
    params: extract.S3Params,
) -> S3SerializationType:
    artifact_type = params.artifact_type
    if artifact_type == ArtifactType.TABLE:
        if params.format is None:
            raise Exception("You must specify a file format for table data.")

        if params.format == common.S3TableFormat.CSV:
            return S3SerializationType.CSV_TABLE
        elif params.format == common.S3TableFormat.JSON:
            return S3SerializationType.JSON_TABLE
        elif params.format == common.S3TableFormat.PARQUET:
            return S3SerializationType.PARQUET_TABLE
        else:
            raise Exception(
                "Unknown S3 file format `%s` for file at path `%s`." % (params.format, key)
            )
    elif artifact_type == ArtifactType.JSON:
        serialization_type = S3SerializationType.JSON
    elif artifact_type == ArtifactType.IMAGE:
        serialization_type = S3SerializationType.IMAGE
    elif (
        artifact_type == ArtifactType.STRING
        or artifact_type == ArtifactType.BOOL
        or artifact_type == ArtifactType.NUMERIC
        or artifact_type == ArtifactType.DICT
        or artifact_type == ArtifactType.LIST
        or artifact_type == ArtifactType.TUPLE
        or artifact_type == ArtifactType.PICKLABLE
    ):
        serialization_type = S3SerializationType.PICKLE
    else:
        raise Exception(
            "Unsupported data type %s when fetching file at %s." % (params.artifact_type, key)
        )

    assert serialization_type is not None, (
        "Unimplemented case for artifact type `%s`" % artifact_type
    )
    return serialization_type


def _read_csv_table(content: bytes) -> Any:
    return pd.read_csv(io.BytesIO(content))


def _read_json_table(content: bytes) -> Any:
    return pd.read_json(io.BytesIO(content))


def _read_parquet_table(content: bytes) -> Any:
    return pd.read_parquet(io.BytesIO(content))


def _read_json_content(content: bytes) -> Any:
    # This assumes that the encoding is "utf-8". May worth considering letting the user
    # specify custom encoding in the future.
    json_data = content.decode(_DEFAULT_JSON_ENCODING)
    # Make sure the data is a valid json object.
    json.loads(json_data)
    return json_data


def _read_image_content(content: bytes) -> Any:
    return Image.open(io.BytesIO(content))


def _read_pickle_content(content: bytes) -> Any:
    return pickle.loads(content)


_s3_deserialization_function_mapping: Dict[str, Callable[[bytes], Any]] = {
    S3SerializationType.CSV_TABLE: _read_csv_table,
    S3SerializationType.JSON_TABLE: _read_json_table,
    S3SerializationType.PARQUET_TABLE: _read_parquet_table,
    S3SerializationType.JSON: _read_json_content,
    S3SerializationType.IMAGE: _read_json_content,
    S3SerializationType.PICKLE: _read_pickle_content,
}
