import os
from typing import List

import pytest
import utils
from aqueduct.constants.enums import ServiceType
from aqueduct.models.dag import DAG, Metadata
from utils import delete_flow, flow_name_to_id, generate_new_flow_name
from validator import Validator

import aqueduct
from aqueduct import Client, globals


def pytest_addoption(parser):
    # We currently only support a single data integration and compute engine per test suite run.
    parser.addoption(f"--data", action="store", default="aqueduct_demo")
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


API_KEY_ENV_NAME = "API_KEY"
SERVER_ADDR_ENV_NAME = "SERVER_ADDRESS"


@pytest.fixture(scope="function")
def client(pytestconfig):
    # Reset the global dag variable, in case it was dirtied by a previous test,
    # since the dag is a global variable on the aqueduct package.
    globals.__GLOBAL_DAG__ = DAG(metadata=Metadata())

    api_key = os.getenv(API_KEY_ENV_NAME)
    server_address = os.getenv(SERVER_ADDR_ENV_NAME)
    if api_key is None or server_address is None:
        raise Exception(
            "Test Setup Error: api_key and server_address must be set as environmental variables."
        )

    return aqueduct.Client(api_key, server_address)


@pytest.fixture(scope="function")
def data_integration(pytestconfig, client):
    return client.integration(pytestconfig.getoption("data"))


@pytest.fixture(scope="session")
def engine(pytestconfig):
    return pytestconfig.getoption("engine")


@pytest.fixture(autouse=True, scope="session")
def use_deprecated(pytestconfig):
    utils.use_deprecated_code_paths = pytestconfig.getoption("deprecated")


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


@pytest.fixture(autouse=True)
def enable_only_for_data_integration_type(request, client, data_integration):
    """When a test is marked with this, it is enabled for particular ServiceType(s)!

    Eg.
    @pytest.mark.enable_only_for_data_integration_type(*relational_dbs())
    def test_relational_data_integrations_only(data_integration):
        ...
    """
    if request.node.get_closest_marker("enable_only_for_data_integration_type"):
        enabled_data_integration_types = request.node.get_closest_marker(
            "enable_only_for_data_integration_type"
        ).args
        assert all(
            isinstance(data_type, ServiceType) for data_type in enabled_data_integration_types
        ), "Arguments to `enable_only_for_data_integration_type()` must be of type ServiceType"

        if data_integration._metadata.service not in enabled_data_integration_types:
            pytest.skip(
                "Skipped for data integration `%s`, since it is not of type `%s`."
                % (data_integration._metadata.name, ",".join(enabled_data_integration_types))
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
                delete_flow(client, flow_name_to_id[flow_name])

    request.addfinalizer(cleanup_flows)
    return get_new_flow_name


@pytest.fixture(scope="function")
def validator(client, data_integration):
    integration = client.integration(data_integration)
    return Validator(client, integration)
