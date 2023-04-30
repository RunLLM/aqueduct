import pytest
from aqueduct.constants.enums import ServiceType

import aqueduct as aq

from ..setup_integration import get_aqueduct_config
from ..shared.flow_helpers import delete_all_flows
from .data_validator import DataValidator


@pytest.fixture(scope="function")
def data_validator(client, data_integration):
    return DataValidator(client, data_integration)


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

        if data_integration.type() not in enabled_data_integration_types:
            pytest.skip(
                "Skipped for data integration `%s`, since it is not of type `%s`."
                % (data_integration.name(), ",".join(enabled_data_integration_types))
            )


def pytest_sessionfinish(session, exitstatus):
    # hasattr(session.config, "workerinput") ensures
    # this only triggers after all workflow finishes.
    if not hasattr(session.config, "workerinput") and not session.config.getoption("keep_flows"):
        client = aq.Client(*get_aqueduct_config())
        delete_all_flows(client)
