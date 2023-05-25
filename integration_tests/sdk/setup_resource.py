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
from aqueduct.models.resource import BaseResource
from aqueduct.resources.connect_config import AWSCredentialType
from aqueduct.resources.mongodb import MongoDBResource
from aqueduct.resources.s3 import S3Resource
from aqueduct.resources.sql import RelationalDBResource

from aqueduct import Client, get_apikey
from sdk.aqueduct_tests.save import save
from sdk.shared.demo_db import demo_db_tables
from sdk.shared.flow_helpers import delete_flow, publish_flow_test
from sdk.shared.naming import generate_object_name
from sdk.shared.relational import format_table_name

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


def _fetch_demo_data(demo: RelationalDBResource, table_name: str) -> pd.DataFrame:
    df = demo.table(table_name)

    # Certain tables in our demo db read out some unexpected tokens that
    # we need to remove before saving into other databases. The unsanitized version
    # will fail when saved to Snowflake, for example.
    if table_name == "wine" or table_name == "mpg":
        import numpy as np

        df = df.replace(r"^\\N$", np.nan, regex=True)
    return df


def _generate_setup_flow_name(resource: BaseResource):
    return "Setup Data for %s Resource: %s" % (
        resource.type(),
        resource.name(),
    )


def _publish_missing_artifacts(
    client: Client, artifacts: List[BaseArtifact], flow_name: str
) -> None:
    publish_flow_test(
        client,
        artifacts=artifacts,
        name=flow_name,
        engine=None,
    )


def _add_missing_artifacts(
    client: Client,
    resource: BaseResource,
    existing_names: Set[str],
) -> None:
    """Given the names of all objects that already exists in an resource, computes
    any objects that are missing and saves them into the resource.


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

    demo = client.resource("Demo")
    artifacts: List[BaseArtifact] = []
    for table_name in missing_names:
        data = _fetch_demo_data(demo, table_name)
        data_param = client.create_param(generate_object_name(), default=data)

        # We use the generic save() defined in Aqueduct Tests, which dictates
        # the data format.
        save(
            resource,
            cast(TableArtifact, data_param),
            name=format_table_name(table_name, resource.type()),
        )
        artifacts.append(data_param)

    _publish_missing_artifacts(
        client,
        artifacts=artifacts,
        flow_name=_generate_setup_flow_name(resource),
    )


def _setup_mongo_db_data(client: Client, mongo_db: MongoDBResource) -> None:
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


def _setup_external_sqlite_db(path: str):
    """Spins up an external SQLite database at 'path'."""
    assert path[-1] != "/", "Path must point to a file, not a directory."
    import os
    from pathlib import Path

    # Create the parent directories if they don't already exist.
    db_abspath = os.path.expanduser(path)
    db_dirpath = Path((os.path.dirname(db_abspath)))
    db_dirpath.mkdir(parents=True, exist_ok=True)

    # Create the SQLite database.
    _execute_command(["sqlite3", db_abspath, "VACUUM;"])


def _setup_postgres_db():
    _execute_command(["aqueduct", "install", "postgres"])


def _setup_mysql_db():
    _execute_command(["aqueduct", "install", "mysql"])


def _setup_relational_data(client: Client, db: RelationalDBResource) -> None:
    # Find all the tables that already exist.
    existing_table_names = set(db.list_tables()["tablename"])
    _add_missing_artifacts(client, db, existing_table_names)


def _setup_s3_data(client: Client, s3: S3Resource):
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


def setup_data_resources(client: Client, filter_to: Optional[str] = None) -> None:
    """Connects to the given data resource(s) if the server hasn't yet.

    If the data resource is not connected, we ensure that it is spun up properly
    and that the appropriate starting data is populated. We assume that if the resource
    already exists, then all external resources have already been set up.

    If `filter_to` is set, we only connect to that given resource name. Otherwise,
    we attempt to connect to every resource listed in the test config file.
    """
    if filter_to is not None:
        data_resources = [filter_to]
    else:
        data_resources = list_data_resources()

    # No need to do any setup for the demo db.
    if "aqueduct_demo" in data_resources:
        data_resources.remove("aqueduct_demo")
    if "Demo" in data_resources:
        data_resources.remove("Demo")

    if len(data_resources) == 0:
        return

    connected_resources = client.list_resources()
    for resource_name in data_resources:
        # Only connect to resources that don't already exist.
        if resource_name not in connected_resources.keys():
            print(f"Connecting to {resource_name}")
            resource_config = _fetch_resource_credentials("data", resource_name)

            # Stand up the external resource first.
            if resource_config["type"] == ServiceType.SQLITE:
                _setup_external_sqlite_db(resource_config["database"])
            elif resource_config["type"] == ServiceType.POSTGRES:
                _setup_postgres_db()
            elif (
                resource_config["type"] == ServiceType.MYSQL
                or resource_config["type"] == ServiceType.MARIADB
            ):
                _setup_mysql_db()

            client.connect_resource(
                resource_name,
                resource_config["type"],
                _sanitize_resource_config_for_connect(resource_config),
            )

        # Setup the data in each of these resources.
        resource = client.resource(resource_name)
        if isinstance(resource, RelationalDBResource):
            _setup_relational_data(client, resource)
        elif resource.type() == ServiceType.S3:
            _setup_s3_data(client, resource)
        elif resource.type() == ServiceType.MONGO_DB:
            _setup_mongo_db_data(client, resource)
        elif resource.type() == ServiceType.ATHENA:
            # We only support reading from Athena, so no setup is necessary.
            pass
        else:
            raise Exception("Test suite does not yet support %s." % resource.type())


def setup_compute_resources(client: Client, filter_to: Optional[str] = None) -> None:
    """Connects to the given compute resource(s) if the server hasn't yet. It *does not*
    ensure that the compute resources are set up appropriately.

    If `filter_to` is set, we only connect to that given resource name. Otherwise,
    we attempt to connect to every resource listed in the test config file.
    """
    if filter_to is not None:
        compute_resources = [filter_to]
    else:
        compute_resources = list_compute_resources()

    if len(compute_resources) == 0:
        return

    connected_resources = client.list_resources()
    for resource_key in compute_resources:
        if resource_key == "aqueduct_engine":
            # Connect to conda if specified, otherwise, do nothing for aq engine.
            aq_config = _parse_config_file()["compute"][resource_key]
            if aq_config and "conda" in aq_config:
                resource_name = aq_config["conda"]
                if resource_name not in connected_resources.keys():
                    client.connect_resource(
                        resource_name,
                        ServiceType.CONDA,
                        {},  # resource_config
                    )
                    wait_for_conda_resource(client, resource_name)
        # Only connect to resources that don't already exist.
        elif resource_key not in connected_resources.keys():
            resource_name = resource_key
            print(f"Connecting to {resource_name}")
            resource_config = _fetch_resource_credentials("compute", resource_name)

            client.connect_resource(
                resource_name,
                resource_config["type"],
                _sanitize_resource_config_for_connect(resource_config),
            )


def wait_for_conda_resource(client: Client, name: str):
    # Try to preview a test function resource it completes successfully.
    from aqueduct import op

    @op(requirements=["pytest"])
    def test_conda_fn():
        return 123

    while True:
        try:
            _ = test_conda_fn()
            return
        except Exception as e:
            # Throw if error message is not expected.
            if "We are still creating base conda environments" not in str(e):
                raise e

            # Wait and try again if error message is expected.
            time.sleep(5)


def setup_storage_layer(client: Client) -> None:
    """If a storage data resource is specified, perform a migration if we aren't already connected to it."""
    name = get_artifact_store_name()
    if name is None:
        return

    connected_resources = client.list_resources()
    if name not in connected_resources.keys():
        resource_config = _fetch_resource_credentials("data", name)
        resource_config["use_as_storage"] = "true"

        # There is a naming collision between the "type" field in `test-credentials.yml`
        # and the "type" field on the S3Config.
        service_type = resource_config["type"]
        resource_config["type"] = AWSCredentialType.CONFIG_FILE_PATH

        client.connect_resource(
            name,
            service_type,
            resource_config,
        )

        # Poll on the server until the resource is ready.
        while True:
            try:
                _ = client.resource(name)
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


def _sanitize_resource_config_for_connect(config: Dict[str, Any]) -> Dict[str, Any]:
    """WARNING: this modifies the configuration dict."""
    del config["type"]
    return config


def _fetch_resource_credentials(section: str, name: str) -> Dict[str, Any]:
    """
    `section` can be "data" or "compute".
    """
    test_credentials = _parse_credentials_file()
    assert section in test_credentials, "%s section expected in test-credentials.yml" % section

    assert (
        name in test_credentials[section]
    ), "%s Resource `%s` must have its credentials in test-credentials.yml." % (
        section,
        name,
    )
    return test_credentials[section][name]


def is_global_engine_set(name: str) -> bool:
    """
    Returns whether or not the provided compute resource has `set_global_engine` set.

    If name is None (meaning we are using the Aqueduct Server), we return False.
    """
    if not name:
        return False

    test_config = _parse_credentials_file()

    assert "compute" in test_config, "compute section expected in test-config.yml"
    assert name in test_config["compute"].keys(), "%s not in test-config.yml." % name

    if not isinstance(test_config["compute"][name], Dict):
        return False

    return "set_global_engine" in test_config["compute"][name].keys()


def is_lazy_set(name: str) -> bool:
    """
    Returns whether or not the provided compute resource has `set_global_lazy` set.

    If name is None (meaning we are using the Aqueduct Server), we return False.
    """
    if not name:
        return False

    test_config = _parse_config_file()

    assert "compute" in test_config, "compute section expected in test-config.yml"
    assert name in test_config["compute"].keys(), "%s not in test-config.yml." % name

    if not isinstance(test_config["compute"][name], Dict):
        return False

    return "set_global_lazy" in test_config["compute"][name].keys()


def list_data_resources() -> List[str]:
    """Get the list of data resources from the config file."""
    test_config = _parse_config_file()

    assert "data" in test_config, "test-config.yml must have a data section."
    return list(test_config["data"].keys())


def list_compute_resources() -> List[str]:
    """Get the list of compute resources from the config file."""
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


def get_artifact_store_name() -> Optional[str]:
    """Returns None if the artifact store is the local filesystem."""
    test_config = _parse_config_file()
    if "storage" not in test_config:
        return None

    assert len(test_config["storage"]) == 1, "Only one data resource can be set as storage layer."
    return list(test_config["storage"].keys())[0]
