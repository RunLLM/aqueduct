import os
import urllib.parse
import uuid
from typing import Union

import boto3
from aqueduct_executor.operators.connectors.data.config import (
    AthenaConfig,
    AWSCredentialType,
    S3Config,
)


def _session_from_config_file_path(config: Union[S3Config, AthenaConfig]) -> boto3.session.Session:
    """
    returns a boto session that sets access_key and secret_access_key
    based on credentials in config_file_path and config_file_profile.
    """
    os.environ["AWS_SHARED_CREDENTIALS_FILE"] = config.config_file_path
    os.environ["AWS_CONFIG_FILE"] = config.config_file_path
    session = boto3.Session(profile_name=config.config_file_profile)
    # This ensures the credential is cached by the session in memory, so even if the temp credential
    # file is removed, we can still access AWS resources.
    session.get_credentials()
    return session


def _session_from_config_file_content(
    config: Union[S3Config, AthenaConfig]
) -> boto3.session.Session:
    """
    returns a boto session that sets access_key and secret_access_key
    based on credentials in config_file_content and config_file_profile.
    """
    # write to temp file assuming the cwd is safe to create such file.
    temp_path = os.path.join(os.getcwd(), str(uuid.uuid4()))
    with open(temp_path, "w") as w:
        w.write(config.config_file_content)
    config.config_file_path = temp_path

    try:
        return _session_from_config_file_path(config)
    finally:
        # always remove the temp file.
        os.remove(temp_path)


def _session_from_access_key(config: Union[S3Config, AthenaConfig]) -> boto3.session.Session:
    if config.region == "":
        # This is a defensive fallback in case the region field of `S3Config` is empty.
        return boto3.Session(
            aws_access_key_id=config.access_key_id,
            aws_secret_access_key=config.secret_access_key,
        )
    else:
        return boto3.Session(
            aws_access_key_id=config.access_key_id,
            aws_secret_access_key=config.secret_access_key,
            region_name=config.region,
        )


def construct_boto_session(config: Union[S3Config, AthenaConfig]) -> boto3.session.Session:
    if config.type == AWSCredentialType.CONFIG_FILE_CONTENT:
        # Write a temp file
        return _session_from_config_file_content(config)
    elif config.type == AWSCredentialType.CONFIG_FILE_PATH:
        return _session_from_config_file_path(config)
    elif config.type == AWSCredentialType.ACCESS_KEY:
        return _session_from_access_key(config)
    else:
        raise Exception("Unsupported integration config type: %s" % config.type)


def url_encode(value: str) -> str:
    return urllib.parse.quote_plus(value)
