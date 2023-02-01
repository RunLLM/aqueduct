import io
import json
from typing import Any, Dict, List, Optional

import cloudpickle as pickle
import numpy as np
import pandas as pd
from aqueduct_executor.operators.connectors.data import common, connector, extract, load
from aqueduct_executor.operators.connectors.data.config import S3Config
from aqueduct_executor.operators.connectors.data.utils import construct_boto_session
from aqueduct_executor.operators.utils.enums import ArtifactType
from aqueduct_executor.operators.utils.saved_object_delete import SavedObjectDelete
from aqueduct_executor.operators.utils.utils import delete_object
from botocore.client import ClientError
from PIL import Image

_DEFAULT_JSON_ENCODING = "utf8"
_DEFAULT_IMAGE_FORMAT = "jpeg"


class S3Connector(connector.DataConnector):
    def __init__(self, config: S3Config):
        session = construct_boto_session(config)
        self.s3 = session.resource("s3")
        self.bucket = config.bucket

    def authenticate(self) -> None:
        try:
            # Below is a low-overhead way of checking if the user has access to the bucket.
            # Source: https://stackoverflow.com/a/49817544
            self.s3.meta.client.head_bucket(Bucket=self.bucket)
        except ClientError as e:
            raise Exception(
                "Bucket does not exist or you do not have permission to access the bucket: %s."
                % str(e)
            )

    def discover(self) -> List[str]:
        raise Exception("Discover is not supported for S3.")

    def fetch_object(self, key: str, params: extract.S3Params) -> Any:
        response = self.s3.Object(self.bucket, key).get()
        data = response["Body"].read()
        if params.artifact_type == ArtifactType.TABLE:
            if params.format is None:
                raise Exception("You must specify a file format for table data.")
            buf = io.BytesIO(data)

            try:
                if params.format == common.S3TableFormat.CSV:
                    return pd.read_csv(buf)
                elif params.format == common.S3TableFormat.JSON:
                    return pd.read_json(buf)
                elif params.format == common.S3TableFormat.PARQUET:
                    return pd.read_parquet(buf)
            except Exception:
                raise Exception(
                    "Unable to read in table at path `%s` with S3 file format `%s`."
                    % (key, params.format)
                )
            else:
                raise Exception(
                    "Unknown S3 file format `%s` for file at path `%s`." % (params.format, key)
                )

        elif params.artifact_type == ArtifactType.JSON:
            # This assumes that the encoding is "utf-8". May worth considering letting the user
            # specify custom encoding in the future.
            json_data = data.decode(_DEFAULT_JSON_ENCODING)
            # Make sure the data is a valid json object.
            try:
                json.loads(json_data)
                return json_data
            except:
                raise Exception("The file at path `%s` is not a valid JSON object.", key)
        elif params.artifact_type == ArtifactType.IMAGE:
            try:
                return Image.open(io.BytesIO(data))
            except:
                raise Exception("The file at path `%s` is not a valid image object.", key)

        elif params.artifact_type == ArtifactType.BYTES:
            return data
        elif (
            params.artifact_type == ArtifactType.STRING
            or params.artifact_type == ArtifactType.BOOL
            or params.artifact_type == ArtifactType.NUMERIC
            or params.artifact_type == ArtifactType.DICT
            or params.artifact_type == ArtifactType.LIST
            or params.artifact_type == ArtifactType.TUPLE
            or params.artifact_type == ArtifactType.PICKLABLE
        ):
            try:
                unpickled_data = pickle.loads(data)
            except Exception:
                raise Exception(
                    "The file at path `%s` is not a valid %s object." % (key, params.artifact_type),
                )

            if params.artifact_type == ArtifactType.STRING:
                if not isinstance(unpickled_data, str):
                    raise Exception(
                        "The file at path `%s` is expected to be a string, got %s."
                        % (key, type(unpickled_data))
                    )
            elif params.artifact_type == ArtifactType.BOOL:
                if not (isinstance(unpickled_data, bool) or isinstance(unpickled_data, np.bool_)):
                    raise Exception(
                        "The file at path `%s` is expected to be a bool, got %s."
                        % (key, type(unpickled_data))
                    )
            elif params.artifact_type == ArtifactType.NUMERIC:
                if not (
                    isinstance(unpickled_data, int)
                    or isinstance(unpickled_data, float)
                    or isinstance(unpickled_data, np.number)
                ):
                    raise Exception(
                        "The file at path `%s` is expected to be a numeric, got %s."
                        % (key, type(unpickled_data))
                    )
            elif params.artifact_type == ArtifactType.DICT:
                if not isinstance(unpickled_data, dict):
                    raise Exception(
                        "The file at path `%s` is expected to be a dictionary, got %s."
                        % (key, type(unpickled_data))
                    )
            elif params.artifact_type == ArtifactType.TUPLE:
                if not isinstance(unpickled_data, tuple):
                    raise Exception(
                        "The file at path `%s` is expected to be a tuple, got %s."
                        % (key, type(unpickled_data))
                    )

            return unpickled_data
        else:
            raise Exception(
                "Unsupported data type %s when fetching file at %s." % (params.artifact_type, key)
            )

    def extract(self, params: extract.S3Params) -> Any:
        path = json.loads(params.filepath)
        if not isinstance(path, List):
            if len(path) == 0:
                raise Exception("S3 file path cannot be an empty string.")
            if path[-1] == "/":
                # This means the path is a directory, and we will do a prefix search.
                files = []
                for obj in self.s3.Bucket(self.bucket).objects.filter(Prefix=path):
                    # The filter api also returns the directories, so we filter them out.
                    if (obj.key)[-1] != "/":
                        files.append(self.fetch_object(obj.key, params))

                if params.artifact_type == ArtifactType.TABLE and params.merge:
                    # We ignore indexes anyways when serializing the data later, so it's ok to do it earlier here.
                    return pd.concat(files, ignore_index=True)
                else:
                    return tuple(files)
            else:
                # This means the path is a file name, and we do a regular file retrieval.
                return self.fetch_object(path, params)
        else:
            # This means we have a list of file paths.
            files = []
            for key in path:
                if len(key) == 0:
                    raise Exception("S3 file path cannot be an empty string.")
                if key[-1] == "/":
                    raise Exception("Each key in the list must not be a directory, found %s." % key)
                files.append(self.fetch_object(key, params))

            if params.artifact_type == ArtifactType.TABLE and params.merge:
                # We ignore indexes anyways when serializing the data later, so it's ok to do it earlier here.
                return pd.concat(files, ignore_index=True)
            else:
                return tuple(files)

    def load(self, params: load.S3Params, data: Any, artifact_type: ArtifactType) -> None:
        if artifact_type == ArtifactType.TABLE:
            if params.format is None:
                raise Exception("You must specify a file format for table data.")
            buf = io.BytesIO()
            if params.format == common.S3TableFormat.CSV:
                data.to_csv(buf, index=False)
            elif params.format == common.S3TableFormat.JSON:
                # Index cannot be False for `to.json` for default orient
                # See: https://pandas.pydata.org/docs/reference/api/pandas.DataFrame.to_json.html
                data.to_json(buf)
            elif params.format == common.S3TableFormat.PARQUET:
                data.to_parquet(buf, index=False)
            else:
                raise Exception("Unknown S3 file format %s." % params.format)
            serialized_data = buf.getvalue()
        elif artifact_type == ArtifactType.JSON:
            serialized_data = data.encode(_DEFAULT_JSON_ENCODING)
        elif artifact_type == ArtifactType.IMAGE:
            img_bytes = io.BytesIO()
            data.save(img_bytes, format=_DEFAULT_IMAGE_FORMAT)
            serialized_data = img_bytes.getvalue()
        elif artifact_type == ArtifactType.BYTES:
            serialized_data = data
        elif (
            artifact_type == ArtifactType.STRING
            or artifact_type == ArtifactType.BOOL
            or artifact_type == ArtifactType.NUMERIC
            or artifact_type == ArtifactType.DICT
            or artifact_type == ArtifactType.LIST
            or artifact_type == ArtifactType.TUPLE
            or artifact_type == ArtifactType.PICKLABLE
        ):
            serialized_data = pickle.dumps(data)
        else:
            raise Exception("Unsupported data type `%s`." % artifact_type)

        self.s3.Object(self.bucket, params.filepath).put(Body=serialized_data)

    def _delete_object(self, name: str, context: Optional[Dict[str, Any]] = None) -> None:
        """`name` is expected to be the S3 FilePath, found in S3Params."""
        self.s3.Object(self.bucket, name).delete()

    def delete(self, objects: List[str]) -> List[SavedObjectDelete]:
        results = []
        for key in objects:
            results.append(delete_object(key, self._delete_object))
        return results
