import pytest

def pytest_addoption(parser):
    parser.addoption("--apikey", type=str, required=True)
    parser.addoption("--address", type=str, default="http://localhost:8080")

def pytest_configure(config):
    pytest.apikey = config.getoption('apikey')
    pytest.address = config.getoption('address')