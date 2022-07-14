import pytest
import os
import aqueduct

API_KEY_ENV_NAME = "API_KEY"
SERVER_ADDR_ENV_NAME = "SERVER_ADDRESS"

def pytest_configure(config):
    pytest.apikey = os.getenv(API_KEY_ENV_NAME)
    pytest.server_address = os.getenv(SERVER_ADDR_ENV_NAME)