from typing import Any, Dict, List, Optional, Set, Tuple, cast

import pandas as pd
import yaml
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.table_artifact import TableArtifact
from aqueduct.constants.enums import ArtifactType, ServiceType
from aqueduct.integrations.s3_integration import S3Integration
from aqueduct.integrations.sql_integration import RelationalDBIntegration
from aqueduct.models.integration import Integration

from aqueduct import Client
from sdk.aqueduct_tests.save import save
from sdk.shared.demo_db import demo_db_tables
from sdk.shared.flow_helpers import delete_flow, publish_flow_test
from sdk.shared.naming import generate_object_name

TEST_CONFIG_FILE: str = "test-config.yml"

# We only cache the config for the lifecycle of a single test run.
CACHED_CONFIG: Optional[Dict[str, Any]] = None


def _parse_config_file() -> Dict[str, Any]:
    global CACHED_CONFIG
    if CACHED_CONFIG is None:
        with open(TEST_CONFIG_FILE, "r") as f:
            CACHED_CONFIG = yaml.safe_load(f)

    return CACHED_CONFIG


def _fetch_demo_data(demo: RelationalDBIntegration, table_name: str) -> pd.DataFrame:
    df = demo.table(table_name)

    # Certain tables in our demo db read out some unexpected tokens that
    # we need to remove before saving into other databases. The unsanitized version
    # will fail when saved to Snowflake, for example.
    if table_name == "wine" or table_name == "mpg":
        import numpy as np

        df = df.replace(r"^\\N$", np.nan, regex=True)
    return df


def _generate_setup_flow_name(integration: Integration):
    return "Setup Data for %s Integration: %s" % (
        integration.type(),
        integration.name(),
    )


def _publish_missing_artifacts(
    client: Client, artifacts: List[BaseArtifact], flow_name: str
) -> None:
    flow = publish_flow_test(
        client,
        artifacts=artifacts,
        name=flow_name,
        engine=None,
    )
    delete_flow(client, flow.id())


def _add_missing_artifacts(
    client: Client,
    integration: Integration,
    existing_names: Set[str],
) -> None:
    """Given the names of all objects that already exists in an integration, computes
    any objects that are missing and saves them into the integration.


    Publishes a workflow in order to do this. The workflow is immediately deleted after one
    successful run.  All setup data is extracted from the demo db.
    """
    # Force name comparisons to be case-insensitive.
    existing_names = [elem.lower() for elem in existing_names]

    needed_names = set(demo_db_tables())
    already_set_up_names = needed_names.intersection(existing_names)
    missing_names = needed_names.difference(already_set_up_names)
    if len(missing_names) == 0:
        return

    demo = client.integration("aqueduct_demo")
    artifacts: List[BaseArtifact] = []
    for table_name in missing_names:
        data = _fetch_demo_data(demo, table_name)
        data_param = client.create_param(generate_object_name(), default=data)

        # We use the generic save() defined in Aqueduct Tests, which dictates
        # the data format.
        save(
            integration,
            cast(TableArtifact, data_param),
            name=table_name,
        )
        artifacts.append(data_param)

    _publish_missing_artifacts(
        client,
        artifacts=artifacts,
        flow_name=_generate_setup_flow_name(integration),
    )


def _setup_relational_data(client: Client, relationalDB: RelationalDBIntegration) -> None:
    # Find all the tables that already exist.
    existing_table_names = set(relationalDB.list_tables()["tablename"])

    _add_missing_artifacts(client, relationalDB, existing_table_names)


def _setup_s3_data(client: Client, s3: S3Integration):

    # Find all the objects that already exist.
    existing_names = set()
    for object_name in demo_db_tables():
        try:
            s3.file(object_name, artifact_type=ArtifactType.TABLE, format="parquet")
            existing_names.add(object_name)
        except:
            # Failing to fetch simply means we will need to populate this data.
            pass

    _add_missing_artifacts(client, s3, existing_names)


def setup_data_integrations(filter_to: Optional[str] = None) -> None:
    """Connects to the given integration name if the server hasn't yet. It also ensures
    that the appropriate data is populated.

    If `filter_to` is set, we only connect to that given integration name. Otherwise,
    we attempt to connect to every integration listed in the config file.
    """
    if filter_to is not None:
        data_integrations = [filter_to]
    else:
        data_integrations = list_data_integrations()

    # No need to do any setup for the demo db.
    if "aqueduct_demo" in data_integrations:
        data_integrations.remove("aqueduct_demo")

    if len(data_integrations) == 0:
        return

    test_config = _parse_config_file()
    assert "data" in test_config

    client = Client(*get_aqueduct_config())
    for integration_name in data_integrations:
        connected_integrations = client.list_integrations()

        # Only connect to integrations that don't already exist.
        if integration_name not in connected_integrations.keys():
            assert integration_name in test_config["data"], (
                "Data integration `%s` needs to exist in the test configuration file."
                % integration_name
            )

            integration_config = test_config["data"][integration_name]
            service_type = integration_config["type"]

            # Modifying the config dictionary should be ok, since we only ever process
            # an entry once.
            del integration_config["type"]
            client.connect_integration(integration_name, service_type, integration_config)

        # Setup the data in each of these integrations.
        integration = client.integration(integration_name)
        if isinstance(integration, RelationalDBIntegration):
            _setup_relational_data(client, integration)
        elif integration.type() == ServiceType.S3:
            _setup_s3_data(client, integration)
        else:
            raise Exception("Test suite does not yet support %s." % integration.type())


def list_data_integrations() -> List[str]:
    """Get the list of data integrations from the config file."""
    test_config = _parse_config_file()

    data_integrations = ["aqueduct_demo"]
    if "data" in test_config:
        data_integrations += list(test_config["data"].keys())
    return data_integrations


def get_aqueduct_config() -> Tuple[str, str]:
    # Returns the apikey and server address.
    test_config = _parse_config_file()
    assert (
        "apikey" in test_config and "address" in test_config
    ), "apikey and address must be set in test-config.yml."
    return test_config["apikey"], test_config["address"]
