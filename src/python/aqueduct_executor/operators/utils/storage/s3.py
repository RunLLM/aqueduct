import os
from typing import Any, Tuple

import boto3
from aqueduct_executor.operators.utils.storage.config import S3StorageConfig
from aqueduct_executor.operators.utils.storage.storage import Storage
from botocore.config import Config as BotoConfig
from botocore.exceptions import ClientError


class S3Storage(Storage):
    _client: Any  # boto3 s3 client
    _config: S3StorageConfig

    def __init__(self, config: S3StorageConfig):
        if config.aws_access_key_id and config.aws_secret_access_key:
            # The AWS keys are passed in as part of the storage spec for AWS Lambda engines
            self._client = boto3.client(
                "s3",
                aws_access_key_id=config.aws_access_key_id,
                aws_secret_access_key=config.aws_secret_access_key,
            )
        elif "AWS_ACCESS_KEY_ID" in os.environ and "AWS_SECRET_ACCESS_KEY" in os.environ:
            # The AWS keys are passed in as environment variables for k8s engines
            self._client = boto3.client(
                "s3",
                aws_access_key_id=os.environ["AWS_ACCESS_KEY_ID"],
                aws_secret_access_key=os.environ["AWS_SECRET_ACCESS_KEY"],
            )
        else:
            # Boto3 uses an environment variable to determine the credentials filepath and profile
            os.environ["AWS_SHARED_CREDENTIALS_FILE"] = config.credentials_path
            os.environ["AWS_PROFILE"] = config.credentials_profile
            self._client = boto3.client("s3", config=BotoConfig(region_name=config.region))

        self._config = config

        bucket, key_prefix = parse_s3_bucket_and_key_prefix(
            self._config.bucket, self._config.root_dir
        )
        self._bucket = bucket
        self._key_prefix = key_prefix

    def put(self, key: str, value: bytes) -> None:
        key = self._resolve_full_key(key)
        print(f"writing to s3: {key}")
        self._client.put_object(Bucket=self._bucket, Key=key, Body=value)

    def get(self, key: str) -> bytes:
        key = self._resolve_full_key(key)
        print(f"reading from s3: {key}")
        return self._client.get_object(Bucket=self._bucket, Key=key)["Body"].read()  # type: ignore

    def exists(self, key: str) -> bool:
        key = self._resolve_full_key(key)
        print(f"checking if exists in s3: {key}")
        try:
            self._client.head_object(Bucket=self._bucket, Key=key)
        except ClientError as e:
            # checking for NoSuchKey error
            if e.response["Error"]["Code"] == "404":
                return False
            else:
                # TODO: ENG-2428 we should explicitly surface other error types to the caller
                # instead of just returning `false` for non s3.ErrCodeNoSuchKey errors.
                return False
        return True

    def _resolve_full_key(self, key: str) -> str:
        """The `key_prefix` is expected to be in the format `path/to/dir/`."""
        if not self._key_prefix:
            return key

        assert self._key_prefix[0] != "/" and self._key_prefix[-1] == "/"
        return self._key_prefix + key


def _sanitize_path(path: str) -> str:
    """Sanitize the given path to be in the format `path/to/dir/` (no leading slash but one trailing slash)."""
    if path == "":
        return path
    if path[0] == "/":
        path = path[1:]
    if path[-1] != "/":
        path += "/"
    return path


def parse_s3_bucket_and_key_prefix(s3_bucket_path: str, root_dir_path: str) -> Tuple[str, str]:
    path_parts = s3_bucket_path.replace("s3://", "").split("/")
    bucket = path_parts.pop(0)

    if root_dir_path != "":
        path_parts += [_sanitize_path(root_dir_path)]

    key_prefix = "/".join(path_parts)
    return bucket, _sanitize_path(key_prefix)
