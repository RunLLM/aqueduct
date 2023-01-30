import functools
import io
import json
from typing import Any, Dict, List, Optional

import cloudpickle as pickle
import numpy as np
import pandas as pd
from aqueduct_executor.operators.connectors.data import common, connector, extract, load, s3
from aqueduct_executor.operators.connectors.data.config import S3Config
from aqueduct_executor.operators.connectors.data.utils import construct_boto_session
from aqueduct_executor.operators.utils.enums import ArtifactType
from aqueduct_executor.operators.utils.saved_object_delete import SavedObjectDelete
from aqueduct_executor.operators.utils.utils import delete_object
from botocore.client import ClientError
from PIL import Image
from pyspark.sql import DataFrame, SparkSession

_DEFAULT_JSON_ENCODING = "utf8"
_DEFAULT_IMAGE_FORMAT = "jpeg"
s3_template = "s3a://%s/%s"


class SparkS3Connector(s3.S3Connector):
    def __init__(self, config: S3Config):

        super().__init__(config)

    def _fetch_object_spark(
        self, key: str, params: extract.S3Params, spark_session_obj: SparkSession
    ) -> Any:
        # Table artifacts use spark to load data into Spark DataFrames.
        if params.artifact_type == ArtifactType.TABLE:
            if params.format is None:
                raise Exception("You must specify a file format for table data.")
            data_path = s3_template % (self.bucket, key)
            try:
                if params.format == common.S3TableFormat.CSV:
                    return spark_session_obj.read.option("header", "true").csv(data_path)
                elif params.format == common.S3TableFormat.JSON:
                    return spark_session_obj.read.json(data_path)
                elif params.format == common.S3TableFormat.PARQUET:
                    return spark_session_obj.read.parquet(data_path)
            except Exception:
                raise Exception(
                    "Unable to read in table at path `%s` with S3 file format `%s`."
                    % (key, params.format)
                )
            else:
                raise Exception(
                    "Unknown S3 file format `%s` for file at path `%s`." % (params.format, key)
                )
        # Non-table artifacts use same serialization as regular S3 integration.
        else:
            return self.fetch_object(key, params)

    def extract_spark(self, params: extract.S3Params, spark_session_obj: SparkSession) -> Any:
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
                        files.append(self._fetch_object_spark(obj.key, params, spark_session_obj))

                if params.artifact_type == ArtifactType.TABLE and params.merge:
                    # We ignore indexes anyways when serializing the data later, so it's ok to do it earlier here.
                    return pd.concat(files, ignore_index=True)
                else:
                    return tuple(files)
            else:
                # This means the path is a file name, and we do a regular file retrieval.
                return self._fetch_object_spark(path, params, spark_session_obj)
        else:
            # This means we have a list of file paths.
            files = []
            for key in path:
                if len(key) == 0:
                    raise Exception("S3 file path cannot be an empty string.")
                if key[-1] == "/":
                    raise Exception("Each key in the list must not be a directory, found %s." % key)
                files.append(self._fetch_object_spark(key, params, spark_session_obj))

            if params.artifact_type == ArtifactType.TABLE and params.merge:
                # We ignore indexes anyways when serializing the data later, so it's ok to do it earlier here.
                return self.unionAll(files)
            else:
                return tuple(files)

    def load_spark(self, params: load.S3Params, data: Any, artifact_type: ArtifactType) -> None:
        if artifact_type == ArtifactType.TABLE:
            if params.format is None:
                raise Exception("You must specify a file format for table data.")

            # data is a Spark DataFrame.
            data_path = s3_template % (self.bucket, params.filepath)
            if params.format == common.S3TableFormat.CSV:
                data.write.csv(data_path)
            elif params.format == common.S3TableFormat.JSON:
                data.write.json(data_path)
            elif params.format == common.S3TableFormat.PARQUET:
                data.write.parquet(data_path)
            else:
                raise Exception("Unknown S3 file format %s." % params.format)
        else:
            # data is not a Spark DataFrame, use normal S3 integration's load.
            self.load(params, data, artifact_type)

    def unionAll(self, dfs: Any) -> Any:
        return functools.reduce(lambda df1, df2: df1.union(df2.select(df1.columns)), dfs)
