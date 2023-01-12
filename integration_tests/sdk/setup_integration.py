from typing import Dict, List, Any, Optional, Tuple

import yaml

from aqueduct import Client
from aqueduct.constants.enums import ServiceType

TEST_CONFIG_FILE: str = "test-config-example.yml"
CACHED_CONFIG: Optional[Dict[str, Any]] = None

def setup_snowflake_data(client):
    pass


def setup_sqlite_data(client):
    pass


def setup_s3_data(client):
    pass


def _parse_config_file() -> Dict[str, Any]:
    global CACHED_CONFIG
    if CACHED_CONFIG is None:
        with open(TEST_CONFIG_FILE, "r") as f:
            CACHED_CONFIG = yaml.safe_load(f)

    return CACHED_CONFIG


# TODO: do not require a full setup of the yaml file if you're just trying to run a single integration.
def setup_data_integrations():
    """Returns the list of data integrations that we expect the tests to run against.

    This list of integrations is configured by `integrations-config.yml`. This method connects
    to any integrations that the server doesn't have, and also ensures that the appropriate data
    is populated in each one.
    """
    test_config = _parse_config_file()

    assert "apikey" in test_config
    assert "address" in test_config
    assert "data" in test_config

    client = Client(test_config["apikey"], test_config["address"])
    connected_integrations = client.list_integrations()
    for name, config in test_config["data"].items():
        service_type = config["type"]

        # Connect to any integrations that don't exist.
        if name not in connected_integrations.keys():
            del config["type"]
            client.connect_integration(name, service_type, config)

        # Setup the data in each of these integrations.
        if service_type == ServiceType.SNOWFLAKE:
            setup_snowflake_data(client)
        elif service_type == ServiceType.SQLITE:
            setup_sqlite_data(client)
        elif service_type == ServiceType.S3:
            setup_s3_data(client)
        else:
            raise Exception("Test suite does not yet support %s." % service_type)


def list_data_integrations() -> List[str]:
    """Assumption is that `setup_data_integrations()` has already run."""
    test_config = _parse_config_file()
    data_integrations = list(test_config["data"].keys())
    data_integrations.insert(0, "aqueduct_demo")
    return data_integrations


def get_aqueduct_config() -> Tuple[str, str]:
    # Returns the apikey and server address.
    test_config = _parse_config_file()
    return test_config["apikey"], test_config["address"]
