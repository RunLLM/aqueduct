import io
import json
from typing import Any, Callable, Dict, List, Optional

import cloudpickle as pickle
import pandas as pd
from aqueduct.constants.enums import S3SerializationType
from aqueduct.utils.serialization import PickleableCollectionSerializationFormat
from aqueduct.utils.type_inference import infer_artifact_type
from aqueduct_executor.operators.connectors.data.common import S3TableFormat
from aqueduct_executor.operators.utils.enums import ArtifactType
from PIL import Image

_DEFAULT_JSON_ENCODING = "utf8"
_DEFAULT_IMAGE_FORMAT = "jpeg"


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


def _read_bytes_content(content: bytes) -> Any:
    return content


def _read_image_content(content: bytes) -> Any:
    return Image.open(io.BytesIO(content))


def _read_pickle_content(content: bytes) -> Any:
    return pickle.loads(content)


_s3_deserialization_function_mapping: Dict[str, Callable[[bytes], Any]] = {
    S3SerializationType.CSV_TABLE: _read_csv_table,
    S3SerializationType.JSON_TABLE: _read_json_table,
    S3SerializationType.PARQUET_TABLE: _read_parquet_table,
    S3SerializationType.JSON: _read_json_content,
    S3SerializationType.BYTES: _read_bytes_content,
    S3SerializationType.IMAGE: _read_image_content,
    S3SerializationType.PICKLE: _read_pickle_content,
}


def _write_csv_table(output: pd.DataFrame) -> bytes:
    buf = io.BytesIO()
    output.to_csv(buf, index=False)
    return buf.getvalue()


def _write_json_table(output: pd.DataFrame) -> bytes:
    buf = io.BytesIO()
    # Index cannot be False for `to.json` for default orient
    # See: https://pandas.pydata.org/docs/reference/api/pandas.DataFrame.to_json.html
    output.to_json(buf)
    return buf.getvalue()


def _write_parquet_table(output: pd.DataFrame) -> bytes:
    buf = io.BytesIO()
    output.to_parquet(buf, index=False)
    return buf.getvalue()


def _write_json_content(output: str) -> bytes:
    return output.encode(_DEFAULT_JSON_ENCODING)


def _write_bytes_content(output: bytes) -> bytes:
    return output


def _write_image_content(output: Image.Image) -> bytes:
    img_bytes = io.BytesIO()
    output.save(img_bytes, format=_DEFAULT_IMAGE_FORMAT)
    return img_bytes.getvalue()


def _write_pickle_content(output: Any) -> bytes:
    return bytes(pickle.dumps(output))


__s3_serialization_function_mapping: Dict[str, Callable[..., bytes]] = {
    S3SerializationType.CSV_TABLE: _write_csv_table,
    S3SerializationType.JSON_TABLE: _write_json_table,
    S3SerializationType.PARQUET_TABLE: _write_parquet_table,
    S3SerializationType.JSON: _write_json_content,
    S3SerializationType.BYTES: _write_bytes_content,
    S3SerializationType.IMAGE: _write_image_content,
    S3SerializationType.PICKLE: _write_pickle_content,
}


def serialize_val_for_s3(
    val: Any,
    serialization_type: S3SerializationType,
    format: Optional[S3TableFormat],
) -> bytes:
    """Serializes a computed value into bytes to be uploaded to S3.

    Unlike deserialization, the serialization of data to S3 relies on potentially
    inferring the artifact and serialization type of elements in a list/tuple. Since
    the artifact -> serialization mapping is different for S3, it's cleaner to just
    implement an S3 version of serialization.
    """
    if serialization_type == S3SerializationType.PICKLE and (
        isinstance(val, list) or isinstance(val, tuple)
    ):
        elem_serialization_types: List[S3SerializationType] = []
        for elem in val:
            elem_artifact_type = infer_artifact_type(elem)

            elem_serialization_types.append(
                artifact_type_to_s3_serialization_type(
                    elem_artifact_type,
                    format,
                    ignore_format_requirement=True,
                )
            )

        data: List[bytes] = [
            __s3_serialization_function_mapping[elem_serialization_types[i]](val[i])
            for i in range(len(elem_serialization_types))
        ]

        pickled_collection_data = PickleableCollectionSerializationFormat(
            aqueduct_serialization_types=elem_serialization_types,
            data=data,
            is_tuple=isinstance(val, tuple),
        )

        # The value we end up pickling is a dictionary.
        val = pickled_collection_data.dict()

    return __s3_serialization_function_mapping[serialization_type](val)


class S3UnsupportedArtifactTypeException(Exception):
    pass


class S3UnknownFileFormatException(Exception):
    pass


class S3InsufficientPermissionsException(Exception):
    pass


class S3RootFolderCreationException(Exception):
    pass


def artifact_type_to_s3_serialization_type(
    artifact_type: ArtifactType,
    format: Optional[S3TableFormat],
    ignore_format_requirement: bool = False,
) -> S3SerializationType:
    if artifact_type == ArtifactType.TABLE:
        if format is None:
            if not ignore_format_requirement:
                raise Exception("You must specify a file format for table data.")

            # Default format is Parquet.
            format = S3TableFormat.PARQUET

        if format == S3TableFormat.CSV:
            return S3SerializationType.CSV_TABLE
        elif format == S3TableFormat.JSON:
            return S3SerializationType.JSON_TABLE
        elif format == S3TableFormat.PARQUET:
            return S3SerializationType.PARQUET_TABLE
        else:
            raise S3UnknownFileFormatException("Unknown S3 file format `%s`" % format)
    elif artifact_type == ArtifactType.JSON:
        serialization_type = S3SerializationType.JSON
    elif artifact_type == ArtifactType.IMAGE:
        serialization_type = S3SerializationType.IMAGE
    elif artifact_type == ArtifactType.BYTES:
        serialization_type = S3SerializationType.BYTES

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
        raise S3UnsupportedArtifactTypeException("Unsupported data type %s." % artifact_type)

    assert serialization_type is not None, (
        "Unimplemented case for artifact type `%s`" % artifact_type
    )
    return serialization_type
