import io
import json
import os
import uuid
from typing import List, Optional

import boto3
import pandas as pd
from aqueduct_executor.operators.connectors.tabular import common, connector, extract, load
from aqueduct_executor.operators.connectors.tabular.config import S3Config, S3CredentialType


class S3Connector(connector.TabularConnector):
    def _handle_config_file_path(self, config: S3Config) -> None:
        """
        _handle_config_file_path updates config's access_key and secret_access_key
        based on credentials in config_file_path and config_file_profile.
        """
        os.environ["AWS_SHARED_CREDENTIALS_FILE"] = config.config_file_path
        os.environ["AWS_CONFIG_FILE"] = config.config_file_path
        session = boto3.Session(profile_name=config.config_file_profile)
        credentials = session.get_credentials()
        config.access_key_id = credentials.access_key
        config.secret_access_key = credentials.secret_key

    def _handle_config_file_content(self, config: S3Config) -> None:
        """
        _handle_config_file_content updates config's access_key and secret_access_key
        based on credentials in config_file_content and config_file_profile.
        """
        # write to temp file assuming the cwd is safe to create such file.
        temp_path = os.path.join(os.getcwd(), str(uuid.uuid4()))
        with open(temp_path, "w") as w:
            w.write(config.config_file_content)
        config.config_file_path = temp_path

        try:
            self._handle_config_file_path(config)
        finally:
            # always remove
            os.remove(temp_path)

    def __init__(self, config: S3Config):
        # Write a temp file
        if config.type == S3CredentialType.CONFIG_FILE_CONTENT:
            self._handle_config_file_content(config)

        if config.type == S3CredentialType.CONFIG_FILE_PATH:
            self._handle_config_file_path(config)

        self.s3 = boto3.resource(
            "s3",
            aws_access_key_id=config.access_key_id,
            aws_secret_access_key=config.secret_access_key,
        )

        self.bucket = config.bucket

    def authenticate(self) -> None:
        bucket = self.s3.Bucket(self.bucket)
        # Below is a low-overhead way of checking if the user has access to the bucket.
        # Source: https://stackoverflow.com/a/49817544
        if not bucket.creation_date:
            raise Exception(
                "Bucket does not exist or you do not have permission to access the bucket."
            )

    def discover(self) -> List[str]:
        raise Exception("Discover is not supported for S3.")

    def _fetch_object(self, key: str, format: common.S3FileFormat) -> pd.DataFrame:
        response = self.s3.Object(self.bucket, key).get()
        data = response["Body"].read()
        buf = io.BytesIO(data)
        if format == common.S3FileFormat.CSV:
            return pd.read_csv(buf)
        elif format == common.S3FileFormat.JSON:
            return pd.read_json(buf)
        elif format == common.S3FileFormat.PARQUET:
            return pd.read_parquet(buf)

        raise Exception("Unknown S3 file format %s." % format)

    def extract(self, params: extract.S3Params) -> pd.DataFrame:
        path = json.loads(params.filepath)
        if not isinstance(path, List):
            if len(path) == 0:
                raise Exception("S3 file path cannot be an empty string.")
            if path[-1] == "/":
                # This means the path is a directory, and we will do a prefix search.
                dfs = []
                for obj in self.s3.Bucket(self.bucket).objects.filter(Prefix=path):
                    # The filter api also returns the directories, so we filter them out.
                    if (obj.key)[-1] != "/":
                        dfs.append(self._fetch_object(obj.key, params.format))
                return pd.concat(dfs)
            else:
                # This means the path is a file name, and we do a regular file retrieval.
                return self._fetch_object(path, params.format)
        else:
            # This means we have a list of file paths.
            dfs = []
            for key in path:
                if len(key) == 0:
                    raise Exception("S3 file path cannot be an empty string.")
                if key[-1] == "/":
                    raise Exception("Each key in the list must not be a directory, found %s." % key)
                dfs.append(self._fetch_object(key, params.format))
            return pd.concat(dfs)

    def load(self, params: load.S3Params, df: pd.DataFrame) -> None:
        buf = io.BytesIO()

        if params.format == common.S3FileFormat.CSV:
            df.to_csv(buf, index=False)
        elif params.format == common.S3FileFormat.JSON:
            # Index cannot be False for `to.json` for default orient
            # See: https://pandas.pydata.org/docs/reference/api/pandas.DataFrame.to_json.html
            df.to_json(buf)
        elif params.format == common.S3FileFormat.PARQUET:
            df.to_parquet(buf, index=False)
        else:
            raise Exception("Unknown S3 file format %s" % format)

        self.s3.Object(self.bucket, params.filepath).put(Body=buf.getvalue())
