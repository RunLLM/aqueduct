from typing import Dict, List, Any, Optional, Tuple

import yaml

from aqueduct import Client
from aqueduct.constants.enums import ServiceType

TEST_CONFIG_FILE: str = "test-config-example.yml"

# We only cache the config for the lifecycle of a single test run.
CACHED_CONFIG: Optional[Dict[str, Any]] = None

# Tracks the integrations that we have already set up for this test run.
ready_integrations: set = set()


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


def setup_data_integration(name: str) -> None:
    """Connects to the given integration name if the server hasn't yet. It also ensures
    that the appropriate data is populated.
    """
    if name in ready_integrations:
        return

    test_config = _parse_config_file()
    assert "data" in test_config
    assert name in test_config["data"], "Supplied integration %s not found in config file." % name

    client = Client(*get_aqueduct_config())
    connected_integrations = client.list_integrations()

    integration_config = test_config["data"][name]
    service_type = integration_config["type"]

    # Connect to any integrations that don't exist.
    if name not in connected_integrations.keys():

        # Modifying the config dictionary should be ok, since we only ever process
        # an entry once.
        del integration_config["type"]
        client.connect_integration(name, service_type, integration_config)

    # Setup the data in each of these integrations.
    if service_type == ServiceType.SNOWFLAKE:
        setup_snowflake_data(client)
    elif service_type == ServiceType.SQLITE:
        setup_sqlite_data(client)
    elif service_type == ServiceType.S3:
        setup_s3_data(client)
    else:
        raise Exception("Test suite does not yet support %s." % service_type)

    ready_integrations.add(name)


def list_data_integrations() -> List[str]:
    """Lists all the data integrations present in the config file. The demo db is always included."""
    test_config = _parse_config_file()
    assert "data" in test_config

    data_integrations = list(test_config["data"].keys())
    data_integrations.insert(0, "aqueduct_demo")
    return data_integrations


def get_aqueduct_config() -> Tuple[str, str]:
    # Returns the apikey and server address.
    test_config = _parse_config_file()
    assert "apikey" in test_config and "address" in test_config, "apikey and address must be set in test-config.yml."
    return test_config["apikey"], test_config["address"]
