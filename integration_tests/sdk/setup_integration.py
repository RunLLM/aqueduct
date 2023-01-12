from typing import Dict, List, Any, Optional, Tuple, Set

import yaml
from aqueduct.models.integration import Integration

from aqueduct import Client
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.constants.enums import ServiceType, LoadUpdateMode
from aqueduct.integrations.sql_integration import RelationalDBIntegration
from sdk.shared.demo_db import demo_db_tables
from sdk.shared.flow_helpers import publish_flow_test, delete_flow
from sdk.shared.naming import generate_new_flow_name

TEST_CONFIG_FILE: str = "test-config-example.yml"

# We only cache the config for the lifecycle of a single test run.
CACHED_CONFIG: Optional[Dict[str, Any]] = None

# Tracks the integrations that we have already set up for this test run.
ready_integrations: Set[str] = set()


def _sanitize_wine_data(df):
    """The wine data in our demo db reads out with some unexpected tokens that
    we need to remove before saving into other databases. The unsanitized version
    will fail when saved to Snowflake, for example.
    """
    import numpy as np
    return df.replace(r'^\\N$', np.nan, regex=True)


def _missing_artifacts(client: Client, existing_names: Set[str]) -> Dict[str, BaseArtifact]:
    """Given the names of all objects that already exists in an integration, computes
    the objects that are missing the returns parameter artifacts with the missing data
    for each of the missing objects.

    All setup data is extracted from the demo db.
    """
    needed_names = set(demo_db_tables())
    already_set_up_names = needed_names.intersection(existing_names)

    missing_names = needed_names.difference(already_set_up_names)
    if len(missing_names) == 0:
        return []

    demo = client.integration("aqueduct_demo")
    artifact_by_name: Dict[str, BaseArtifact] = {}
    for table_name in missing_names:
        data = demo.table(table_name)
        if table_name == "wine":
            data = _sanitize_wine_data(data)

        artifact_by_name[table_name] = client.create_param("Snowflake %s Data" % table_name, default=data)
    return artifact_by_name


def setup_snowflake_data(client: Client, snowflake: Integration) -> None:
    assert isinstance(snowflake, RelationalDBIntegration)

    # Check if these tables already exist.
    existing_tables = set(snowflake.list_tables()["tablename"])

    missing_artifact_by_name = _missing_artifacts(existing_tables)
    if len(missing_artifact_by_name) == 0:
        return

    for table_name, artifact in missing_artifact_by_name:
        snowflake.save(artifact, table_name, LoadUpdateMode.REPLACE)

    flow = publish_flow_test(
        client,
        artifacts=list(missing_artifact_by_name.values()),
        name=generate_new_flow_name(),
        engine=None,
    )
    delete_flow(client, flow.id())


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

    integration = client.integration(name)

    # Setup the data in each of these integrations.
    if service_type == ServiceType.SNOWFLAKE:
        setup_snowflake_data(client, integration)
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
