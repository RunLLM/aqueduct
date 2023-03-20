import subprocess
import sys
import time
from typing import Any, Dict, List, Optional, Set, Tuple, cast

import pandas as pd
import yaml
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.table_artifact import TableArtifact
from aqueduct.constants.enums import ArtifactType, ServiceType
from aqueduct.error import AqueductError
from aqueduct.integrations.connect_config import AWSCredentialType
from aqueduct.integrations.mongodb_integration import MongoDBIntegration
from aqueduct.integrations.s3_integration import S3Integration
from aqueduct.integrations.sql_integration import RelationalDBIntegration
from aqueduct.models.integration import Integration

from aqueduct import Client, get_apikey
from sdk.aqueduct_tests.save import save
from sdk.shared.demo_db import demo_db_tables
from sdk.shared.flow_helpers import delete_flow, publish_flow_test
from sdk.shared.naming import generate_object_name

TEST_CREDENTIALS_FILE: str = "test-credentials.yml"
TEST_CONFIG_FILE: str = "test-config.yml"

# We only cache these files for the lifecycle of a single test run.
CACHED_CREDENTIALS: Optional[Dict[str, Any]] = None
CACHED_CONFIG: Optional[Dict[str, Any]] = None


def _execute_command(args, cwd=None):
    with subprocess.Popen(args, stdout=sys.stdout, stderr=sys.stderr, cwd=cwd) as proc:
        proc.communicate()
        if proc.returncode != 0:
            raise Exception("Error executing command: %s" % args)


def _parse_config_file() -> Dict[str, Any]:
    global CACHED_CONFIG
    if CACHED_CONFIG is None:
        with open(TEST_CONFIG_FILE, "r") as f:
            CACHED_CONFIG = yaml.safe_load(f)

    return CACHED_CONFIG


def _parse_credentials_file() -> Dict[str, Any]:
    global CACHED_CREDENTIALS
    if CACHED_CREDENTIALS is None:
        with open(TEST_CREDENTIALS_FILE) as f:
            CACHED_CREDENTIALS = yaml.safe_load(f)

    return CACHED_CREDENTIALS


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


def _setup_mongo_db_data(client: Client, mongo_db: MongoDBIntegration) -> None:
    # Find all the objects that already exist.
    existing_names = set()
    for object_name in demo_db_tables():
        try:
            data = mongo_db.collection(object_name).find({}).get()
            if len(data) > 0:
                existing_names.add(object_name)
        except Exception:
            # Failing to fetch simply means we will need to populate this data.
            pass

    _add_missing_artifacts(client, mongo_db, existing_names)


def _setup_snowflake_data(client: Client, snowflake: RelationalDBIntegration) -> None:
    # Find all the tables that already exist.
    existing_table_names = set(snowflake.list_tables()["tablename"])

    _add_missing_artifacts(client, snowflake, existing_table_names)


def _setup_external_sqlite_db(path: str):
    """Spins up an external SQLite database at 'path'."""
    assert path[-1] != "/", "Path must point to a file"

    import os
    from pathlib import Path

    # Create the parent directories if they don't already exist.
    db_abspath = os.path.expanduser(path)
    db_dirpath = Path((os.path.dirname(db_abspath)))
    db_dirpath.mkdir(parents=True, exist_ok=True)

    # Create the SQLite database.
    _execute_command(["sqlite3", db_abspath, "VACUUM;"])


def _setup_relational_data(client: Client, db: RelationalDBIntegration) -> None:
    # Find all the tables that already exist.
    existing_table_names = set(db.list_tables()["tablename"])
    _add_missing_artifacts(client, db, existing_table_names)


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


def setup_data_integrations(client: Client, filter_to: Optional[str] = None) -> None:
    """Connects to the given data integration(s) if the server hasn't yet.

    If the data integration is not connected, we ensure that it is spun up properly
    and that the appropriate starting data is populated. We assume that if the integration
    already exists, then all external resources have already been set up.

    If `filter_to` is set, we only connect to that given integration name. Otherwise,
    we attempt to connect to every integration listed in the test config file.
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

    connected_integrations = client.list_integrations()
    for integration_name in data_integrations:
        # Only connect to integrations that don't already exist.
        if integration_name not in connected_integrations.keys():
            integration_config = _fetch_integration_credentials("data", integration_name)

            # Stand up the external integration first.
            if integration_config["type"] == ServiceType.SQLITE:
                _setup_external_sqlite_db(integration_config["database"])

            client.connect_integration(
                integration_name,
                integration_config["type"],
                _sanitize_integration_config_for_connect(integration_config),
            )

        # Setup the data in each of these integrations.
        integration = client.integration(integration_name)
        if isinstance(integration, RelationalDBIntegration):
            _setup_relational_data(client, integration)
        elif integration.type() == ServiceType.S3:
            _setup_s3_data(client, integration)
        elif integration.type() == ServiceType.MONGO_DB:
            _setup_mongo_db_data(client, integration)
        elif integration.type() == ServiceType.ATHENA:
            # We only support reading from Athena, so no setup is necessary.
            pass
        else:
            raise Exception("Test suite does not yet support %s." % integration.type())


def setup_compute_integrations(client: Client, filter_to: Optional[str] = None) -> None:
    """Connects to the given compute integration(s) if the server hasn't yet. It *does not*
    ensure that the compute resources are set up appropriately.

    If `filter_to` is set, we only connect to that given integration name. Otherwise,
    we attempt to connect to every integration listed in the test config file.
    """
    if filter_to is not None:
        compute_integrations = [filter_to]
    else:
        compute_integrations = list_compute_integrations()

    # No need to do any setup for the demo db.
    if "aqueduct_engine" in compute_integrations:
        compute_integrations.remove("aqueduct_engine")

    if len(compute_integrations) == 0:
        return

    connected_integrations = client.list_integrations()
    for integration_name in compute_integrations:
        # Only connect to integrations that don't already exist.
        if integration_name not in connected_integrations.keys():
            integration_config = _fetch_integration_credentials("compute", integration_name)

            client.connect_integration(
                integration_name,
                integration_config["type"],
                _sanitize_integration_config_for_connect(integration_config),
            )


def setup_storage_layer(client: Client) -> None:
    """If a storage data integration is specified, perform a migration if we aren't already connected to it."""
    test_config = _parse_config_file()
    if "storage" not in test_config:
        return

    assert (
        len(test_config["storage"]) == 1
    ), "Only one data integration can be set as storage layer."
    name = list(test_config["storage"].keys())[0]

    connected_integrations = client.list_integrations()
    if name not in connected_integrations.keys():
        integration_config = _fetch_integration_credentials("data", name)
        integration_config["use_as_storage"] = "true"

        # There is a naming collision between the "type" field in `test-credentials.yml`
        # and the "type" field on the S3Config.
        service_type = integration_config["type"]
        integration_config["type"] = AWSCredentialType.CONFIG_FILE_PATH

        client.connect_integration(
            name,
            service_type,
            integration_config,
        )

        # Poll on the server until the integration is ready.
        while True:
            try:
                _ = client.integration(name)
            except AqueductError as e:
                if "The server is currently unavailable due to system maintenance." in str(e):
                    time.sleep(1)
                    continue
                raise
            else:
                break


def has_storage_config() -> bool:
    """Check if the test config file has a storage config section (meaning it is using non-local storage)."""
    test_config = _parse_config_file()
    return "storage" in test_config


def _sanitize_integration_config_for_connect(config: Dict[str, Any]) -> Dict[str, Any]:
    """WARNING: this modifies the configuration dict."""
    del config["type"]
    return config


def _fetch_integration_credentials(section: str, name: str) -> Dict[str, Any]:
    """
    `section` can be "data" or "compute".
    """
    test_credentials = _parse_credentials_file()
    assert section in test_credentials, "%s section expected in test-credentials.yml" % section

    assert (
        name in test_credentials[section]
    ), "%s Integration `%s` must have its credentials in test-credentials.yml." % (section, name)
    return test_credentials[section][name]


def list_data_integrations() -> List[str]:
    """Get the list of data integrations from the config file."""
    test_config = _parse_config_file()

    assert "data" in test_config, "test-config.yml must have a data section."
    return list(test_config["data"].keys())


def list_compute_integrations() -> List[str]:
    """Get the list of compute integrations from the config file."""
    test_config = _parse_config_file()

    assert "compute" in test_config, "test-config.yml must have a compute section."
    return list(test_config["compute"].keys())


def get_aqueduct_config() -> Tuple[str, str]:
    """
    Returns the apikey and server address. If "apikey" does not exist in test-credentials, we will
    assume the server is on the same machine, and will use "aqueduct apikey".
    """
    test_config = _parse_config_file()

    test_credentials = _parse_credentials_file()
    if "apikey" in test_credentials:
        apikey = test_credentials["apikey"]
    else:
        apikey = get_apikey()

    return apikey, test_config["address"]
