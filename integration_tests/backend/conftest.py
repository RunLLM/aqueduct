import os

import pytest

API_KEY_ENV_NAME = "API_KEY"
SERVER_ADDR_ENV_NAME = "SERVER_ADDRESS"
INTEGRATION_ENV_NAME = "INTEGRATION"


def pytest_configure(config):
    pytest.api_key = os.getenv(API_KEY_ENV_NAME)
    pytest.server_address = os.getenv(SERVER_ADDR_ENV_NAME)
    pytest.resource = os.getenv(INTEGRATION_ENV_NAME)

    if pytest.api_key is None or pytest.server_address is None or pytest.resource is None:
        raise Exception(
            "Test Setup Error: API_KEY, INTEGRATION, and SERVER_ADDRESS must be set as environmental variables."
        )
