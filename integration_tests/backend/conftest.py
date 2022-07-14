import os

import pytest

import aqueduct

API_KEY_ENV_NAME = "API_KEY"
SERVER_ADDR_ENV_NAME = "SERVER_ADDRESS"
ADAPTER_ENV_NAME = "ADAPTER"


def pytest_configure(config):
    pytest.api_key = os.getenv(API_KEY_ENV_NAME)
    pytest.server_address = os.getenv(SERVER_ADDR_ENV_NAME)
    pytest.adapter = os.getenv(ADAPTER_ENV_NAME)

    if pytest.api_key is None or pytest.server_address is None or pytest.adapter is None:
        raise Exception(
            "Test Setup Error: api_key, server_address, and adapter must be set as environmental variables."
        )
