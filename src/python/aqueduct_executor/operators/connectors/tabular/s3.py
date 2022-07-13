import io
import json
from typing import List

import boto3
import pandas as pd
from aqueduct_executor.operators.connectors.tabular import common, config, connector, extract, load


class S3Connector(connector.TabularConnector):
    def __init__(self, config: config.S3Config):
        if config.region is None:
            self.s3 = boto3.resource(
                "s3",
                aws_access_key_id=config.access_key_id,
                aws_secret_access_key=config.secret_access_key,
            )
        else:
            self.s3 = boto3.resource(
                "s3",
                aws_access_key_id=config.access_key_id,
                aws_secret_access_key=config.secret_access_key,
                region_name = config.region,
            )

        self.bucket = config.bucket

    def authenticate(self) -> None:
        pass

    def discover(self) -> List[str]:
        raise Exception("Discover is not supported for S3.")

    def _parse_data(self, data: io.BytesIO, format: common.S3FileFormat) -> pd.DataFrame:
        if format == common.S3FileFormat.CSV:
            return pd.read_csv(data)
        elif format == common.S3FileFormat.JSON:
            return pd.read_json(data)
        elif format == common.S3FileFormat.PARQUET:
            return pd.read_parquet(data)

        raise Exception("Unknown S3 file format %s" % format)

    def extract(self, params: extract.S3Params) -> pd.DataFrame:
        paths = json.loads(params.filepath)
        if not isinstance(paths, List):
            paths = [paths]

        bucket_obj = self.s3.Bucket(self.bucket)
        dfs = []
        for path in paths:
            for obj in bucket_obj.objects.filter(Prefix=path):
                response = self.s3.Object(self.bucket, obj.key).get()
                data = response["Body"].read()
                buf = io.BytesIO(data)
                dfs.append(self._parse_data(buf, params.format))

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
