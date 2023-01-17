import pytest
from aqueduct.constants.enums import ServiceType
from aqueduct.models.dag import DAG, Metadata

from aqueduct import Client, globals
from sdk.setup_integration import (
    get_aqueduct_config,
    list_data_integrations,
    setup_data_integrations,
)
from sdk.shared import globals as test_globals
from sdk.shared.utils import delete_flow, generate_new_flow_name
from sdk.shared.validator import Validator


def pytest_addoption(parser):
    parser.addoption(f"--data", action="store", default=None)
    parser.addoption(f"--engine", action="store", default=None)
    parser.addoption(f"--keep-flows", action="store_true", default=False)

    # Sets a global flag that can be toggled if we want to check that a deprecated code path still works.
    parser.addoption(f"--deprecated", action="store_true", default=False)


def pytest_configure(config):
    """This is just to prevent warnings around our custom markers. eg. `pytest.mark.enable_only_for_engine`."""
    config.addinivalue_line(
        "markers", "enable_only_for_engine_type: runs the test only for the supplied engines."
    )
    config.addinivalue_line(
        "markers",
        "enable_only_for_data_integration_type: runs the test only for the supplied data integrations.",
    )


def pytest_cmdline_main(config):
    """Gets all the integrations ready for the tests to run. Should only run once, before we even collect any tests."""
    data_integration = config.getoption(f"--data")
    if data_integration is not None:
        setup_data_integrations(filter_to=data_integration)
    else:
        setup_data_integrations()


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
            pytest.skip("Skipped. Tests are only running against %s." % cmdline_data_flag)

    return client.integration(request.param)


@pytest.fixture(scope="session")
def engine(pytestconfig):
    return pytestconfig.getoption("engine")


@pytest.fixture(autouse=True, scope="session")
def use_deprecated(pytestconfig):
    test_globals.use_deprecated_code_paths = pytestconfig.getoption("deprecated")


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

        integration_info_by_name = client.list_integrations()
        if engine not in integration_info_by_name.keys():
            raise Exception("Server is not connected to integration `%s`." % engine)

        if integration_info_by_name[engine].service not in enabled_engine_types:
            pytest.skip(
                "Skipped for engine integration `%s`, since it is not of type `%s`."
                % (engine, ",".join(enabled_engine_types))
            )


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
                delete_flow(client, test_globals.flow_name_to_id[flow_name])

    request.addfinalizer(cleanup_flows)
    return get_new_flow_name


@pytest.fixture(scope="function")
def validator(client, data_integration):
    return Validator(client, data_integration)
