from typing import Optional

import pytest
from aqueduct.constants.enums import ServiceType
from aqueduct.models.dag import DAG, Metadata

from aqueduct import Client, global_config, globals
from sdk.setup_integration import (
    get_aqueduct_config,
    get_artifact_store_name,
    has_storage_config,
    is_preview_enabled,
    list_compute_integrations,
    list_data_integrations,
    setup_compute_integrations,
    setup_data_integrations,
    setup_storage_layer,
)
from sdk.shared import globals as test_globals
from sdk.shared.utils import generate_new_flow_name
from sdk.shared.validator import Validator


def pytest_addoption(parser):
    parser.addoption(f"--data", action="store", default=None)
    parser.addoption(f"--engine", action="store", default=None)
    parser.addoption(f"--keep-flows", action="store_true", default=False)

    # Sets a global flag that can be toggled if we want to check that a deprecated code path still works.
    parser.addoption(f"--deprecated", action="store_true", default=False)

    # Skips the setup of data/compute integrations for faster testing. Best used as an optimization after first
    # test run of a debugging session.
    parser.addoption(f"--skip-data-setup", action="store_true", default=False)
    parser.addoption(f"--skip-engine-setup", action="store_true", default=False)

    # Allows any tests that rely on a K8s cluster with a GPU setup to run.
    parser.addoption(f"--gpu", action="store_true", default=False)


def pytest_configure(config):
    """This is just to prevent warnings around our custom markers. eg. `pytest.mark.enable_only_for_engine`."""
    config.addinivalue_line(
        "markers", "enable_only_for_engine_type: runs the test only for the supplied engines."
    )
    config.addinivalue_line(
        "markers",
        "enable_only_for_external_compute: runs the test only for external compute engines.",
    )
    config.addinivalue_line(
        "markers",
        "enable_only_for_data_integration_type: runs the test only for the supplied data integrations.",
    )
    config.addinivalue_line(
        "markers",
        "must_have_gpu: the K8s integration is expected to have access to a GPU.",
    )
    config.addinivalue_line(
        "markers",
        "enable_only_for_local_storage: the test is expected to run in an environment with local storage.",
    )


def pytest_cmdline_main(config):
    """Gets all the integrations ready for the tests to run. Should only run once, before we even collect any tests."""
    client = Client(*get_aqueduct_config())
    setup_storage_layer(client)

    _parse_flags_and_setup_data_integrations(config, client)
    _parse_flags_and_setup_compute_integrations(config, client)


def _parse_flags_and_setup_data_integrations(config, client: Client):
    should_skip = config.getoption(f"--skip-data-setup")
    if should_skip:
        return

    data_integration = config.getoption(f"--data")
    if data_integration is not None:
        setup_data_integrations(client, filter_to=data_integration)
    else:
        setup_data_integrations(client)


def _parse_flags_and_setup_compute_integrations(config, client: Client):
    should_skip = config.getoption(f"--skip-engine-setup")
    if should_skip:
        return

    engine = config.getoption(f"--engine")
    if engine is not None:
        setup_compute_integrations(client, filter_to=engine)
    else:
        setup_compute_integrations(client)


@pytest.fixture(scope="function")
def client(pytestconfig):
    # Reset the global dag variable, in case it was dirtied by a previous test,
    # since the dag is a global variable on the aqueduct package.
    globals.__GLOBAL_DAG__ = DAG(metadata=Metadata())
    return Client(*get_aqueduct_config())


@pytest.fixture(scope="function", params=list_data_integrations())
def data_integration(request, pytestconfig, client):
    """This fixture is parameterized to run every test case against every requested data integration.

    The requested data integrations are all in the test configuration file, but can be overwritten
    by the `--data` command line flag.
    """
    cmdline_data_flag = pytestconfig.getoption("data")
    if cmdline_data_flag is not None:
        if request.param != cmdline_data_flag:
            pytest.skip(
                "Skipped. Tests are only running against data integration %s." % cmdline_data_flag
            )

    return client.integration(request.param)


@pytest.fixture(scope="function", params=list_compute_integrations())
def engine(request, pytestconfig):
    cmdline_compute_flag = pytestconfig.getoption("engine")
    if cmdline_compute_flag is not None:
        if request.param != cmdline_compute_flag:
            pytest.skip(
                "Skipped. Tests are only running against compute %s." % cmdline_compute_flag
            )

    # Test cases process the aqueduct engine as None. We do the conversion here
    # because fixture parameters are printed as part of test execution.
    return request.param if request.param != "aqueduct_engine" else None


@pytest.fixture(scope="function", autouse=True)
def set_global_config(engine):
    # If we are using the aqueduct engine (where the engine fixture is None), we
    # assume that previews are enabled and thus don't have to change the existing
    # global_config.
    if engine != None:
        # If we are using an external compute engine, we check if the `enable_previews` tag
        # has been set in test-credentials.yml. If it is, we set lazy execution to False
        # to force the external engine to run previews. If not set (in the case we want to save
        # on costs) we set lazy to True and thus do not run previews unless we force execution
        # via <artifact>.get().
        lazy_config = not is_preview_enabled(engine)
        global_config({"engine": engine, "lazy": lazy_config})

    yield
    # Reset the global_config after the end of the function.
    global_config({"engine": "aqueduct", "lazy": False})


@pytest.fixture(scope="function")
def artifact_store():
    """Is None if local filesystem is being used as the artifact store."""
    return get_artifact_store_name()


@pytest.fixture(autouse=True, scope="session")
def use_deprecated(pytestconfig):
    test_globals.use_deprecated_code_paths = pytestconfig.getoption("deprecated")


def _type_from_engine_name(client, engine: str) -> ServiceType:
    assert engine != "aqueduct_engine"

    integration_info_by_name = client.list_integrations()
    if engine not in integration_info_by_name.keys():
        raise Exception("Server is not connected to integration `%s`." % engine)

    return integration_info_by_name[engine].service


# Pulled from: https://stackoverflow.com/questions/28179026/how-to-skip-a-pytest-using-an-external-fixture
@pytest.fixture(autouse=True)
def enable_only_for_engine_type(request, client, engine):
    """When a test is marked with this, it is enabled for particular ServiceType(s)!

    Eg.
    @pytest.mark.enable_only_for_engine(ServiceType.LAMBDA, ServiceType.K8s)
    def test_k8s(engine):
        ...
    """
    if request.node.get_closest_marker("enable_only_for_engine_type"):
        enabled_engine_types = request.node.get_closest_marker("enable_only_for_engine_type").args
        assert all(
            isinstance(engine_type, ServiceType) for engine_type in enabled_engine_types
        ), "Arguments to `enable_only_for_engine_type()` must be of type ServiceType"

        if engine is None:
            # We run against the default engine only if Aqueduct engine is specifically enabled.
            # eg. `@pytest.mark.enable_only_for_engine(ServiceType.AQUEDUCT_ENGINE, ServiceType.AIRFLOW)
            if ServiceType.AQUEDUCT_ENGINE in enabled_engine_types:
                return
            else:
                pytest.skip(
                    "Skipped. This test only runs on engine type `%s`."
                    % ",".join(enabled_engine_types)
                )

        if _type_from_engine_name(client, engine) not in enabled_engine_types:
            pytest.skip(
                "Skipped for engine integration `%s`, since it is not of type `%s`."
                % (engine, ",".join(enabled_engine_types))
            )


@pytest.fixture(autouse=True)
def skip_for_spark_engines(request, client, engine):
    """When a test is marked with this, we skip if we are using a spark based engine
    (Databricks or Spark)
    """
    if request.node.get_closest_marker("skip_for_spark_engines"):
        if _type_from_engine_name(client, engine) in [ServiceType.DATABRICKS, ServiceType.SPARK]:
            pytest.skip(
                "Skipped for engine integration `%s`, since it is a spark-based engine."
                % engine
            )


@pytest.fixture(autouse=True)
def enable_only_for_local_storage(request, client, engine):
    """When a test is marked with this, we run it only when the local file system is used as storage."""
    if not request.node.get_closest_marker("enable_only_for_local_storage"):
        return

    if has_storage_config():
        pytest.skip("Skipped since the test environment uses non-local storage.")


@pytest.fixture(autouse=True)
def enable_only_for_external_compute(request, client, engine):
    """When a test is marked with this, it will run for all engine types EXCEPT Aqueduct!"""
    if request.node.get_closest_marker("enable_only_for_external_compute"):
        if engine is None:
            pytest.skip("Skipped. This test only runs against external compute integrations.")


@pytest.fixture(autouse=True)
def must_have_gpu(pytestconfig, request, client, engine):
    """When a test is marked with this, all it means that it will only be executed if the --gpu flag is
    passed into command line.

    The user is responsible for supplying a K8s integration with an available GPU.
    """
    if not request.node.get_closest_marker("must_have_gpu"):
        return

    if pytestconfig.getoption("gpu"):
        assert (
            _type_from_engine_name(client, engine) == ServiceType.K8S
        ), "@pytest.mark.must_have_gpu only works with K8s engine!"
    else:
        pytest.skip("Skipped since --gpu flag is not provided")


@pytest.fixture(scope="function")
def flow_name(client, request, pytestconfig):
    """Any flows created by this fixture will be automatically cleaned up at test teardown.

    Note that it returns a method, so it must be used like:

    ```
    def test_foo(flow_name):
        publish_flow(name=flow_name(), ...)
    ```
    """
    flow_names = []

    def get_new_flow_name():
        flow_name = generate_new_flow_name()
        flow_names.append(flow_name)
        return flow_name

    def cleanup_flows():
        if not pytestconfig.getoption("keep_flows"):
            for flow_name in flow_names:
                try:
                    client.delete_flow(
                        flow_name=flow_name,
                        saved_objects_to_delete=client.flow(
                            flow_name=flow_name
                        ).list_saved_objects(),
                    )
                except Exception as e:
                    print("Error deleting workflow %s with exception: %s" % (flow_name, e))
                else:
                    print("Successfully deleted workflow %s" % flow_name)

    request.addfinalizer(cleanup_flows)
    return get_new_flow_name


@pytest.fixture(scope="function")
def validator(client, data_integration):
    return Validator(client, data_integration)


@pytest.fixture(scope="function", autouse=True)
def post_process_reset_execution_mode_to_eager():
    # Pre-processing code
    yield
    # Post-processing code
    global_config({"lazy": False})
