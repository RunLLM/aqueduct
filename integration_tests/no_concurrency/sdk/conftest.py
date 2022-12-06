import os

import pytest

import aqueduct.globals
from aqueduct.dag.dag import DAG, Metadata

import aqueduct

API_KEY_ENV_NAME = "API_KEY"
SERVER_ADDR_ENV_NAME = "SERVER_ADDRESS"


@pytest.fixture(scope="function")
def client(pytestconfig):
    # Reset the global dag variable, in case it was dirtied by a previous test,
    # since the dag is a global variable on the aqueduct package.
    aqueduct.globals.__GLOBAL_DAG__ = DAG(metadata=Metadata())
    api_key = os.getenv(API_KEY_ENV_NAME)
    server_address = os.getenv(SERVER_ADDR_ENV_NAME)
    if api_key is None or server_address is None:
        raise Exception(
            "Test Setup Error: api_key and server_address must be set as environmental variables."
        )

    return aqueduct.Client(api_key, server_address)
