import os
from typing import List

import pytest
from aqueduct.dag import DAG, Metadata

import aqueduct


def pytest_addoption(parser):
    # We currently only support a single data integration and compute engine per test suite run.
    parser.addoption(f"--data", action="store", default="aqueduct_demo")
    parser.addoption(f"--engine", action="store", default=None)


API_KEY_ENV_NAME = "API_KEY"
SERVER_ADDR_ENV_NAME = "SERVER_ADDRESS"


@pytest.fixture(scope="session")
def data_integration(pytestconfig):
    return pytestconfig.getoption("data")


@pytest.fixture(scope="session")
def engine(pytestconfig):
    return pytestconfig.getoption("engine")


@pytest.fixture(scope="function")
def client(pytestconfig):
    # Reset the global dag variable, in case it was dirtied by a previous test,
    # since the dag is a global variable on the aqueduct package.
    aqueduct.dag.__GLOBAL_DAG__ = DAG(metadata=Metadata())
    api_key = os.getenv(API_KEY_ENV_NAME)
    server_address = os.getenv(SERVER_ADDR_ENV_NAME)
    if api_key is None or server_address is None:
        raise Exception(
            "Test Setup Error: api_key and server_address must be set as environmental variables."
        )

    return aqueduct.Client(api_key, server_address)


# Pulled from: https://stackoverflow.com/questions/28179026/how-to-skip-a-pytest-using-an-external-fixture
@pytest.fixture(autouse=True)
def enable_by_engine_type(request, client, engine):
    """When a test is marked with this, it is enabled for particular ServiceType(s)!

    Eg.
    @pytest.mark.enable_only_for_engine(ServiceType.LAMBDA, ServiceType.K8s)
    def test_k8s(engine):
        ...
    """
    if request.node.get_closest_marker("enable_only_for_engine_type"):
        enabled_engine_types = request.node.get_closest_marker("enable_only_for_engine_type").args

        if engine is None:
            pytest.skip(
                "Skipped. This test only runs on engine type `%s`." % ",".join(enabled_engine_types)
            )
            return

        # Get the type of integration that `engine` is, so we know whether to skip.
        integration_info_by_name = client.list_integrations()
        if engine not in integration_info_by_name.keys():
            raise Exception("Server is not connected an integration `%s`." % engine)

        if integration_info_by_name[engine].service not in enabled_engine_types:
            pytest.skip(
                "Skipped on engine `%s`, since it is not of type `%s`."
                % (engine, ",".join(enabled_engine_types))
            )


# TODO: describe
flow_names_by_test = {}


@pytest.fixture(autouse=True)
def flow_cleanup(request):
    yield

    test_name = (request.node.name)
    flow_id = flow_name_by_test.get(test_name)