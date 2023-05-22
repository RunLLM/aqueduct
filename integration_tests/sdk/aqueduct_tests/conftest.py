import pytest
from aqueduct.constants.enums import ServiceType

import aqueduct as aq

from ..setup_resource import get_aqueduct_config
from ..shared.flow_helpers import delete_all_flows
from .data_validator import DataValidator


@pytest.fixture(scope="function")
def data_validator(client, data_resource):
    return DataValidator(client, data_resource)


@pytest.fixture(autouse=True)
def enable_only_for_data_resource_type(request, client, data_resource):
    """When a test is marked with this, it is enabled for particular ServiceType(s)!

    Eg.
    @pytest.mark.enable_only_for_data_resource_type(*relational_dbs())
    def test_relational_data_resources_only(data_resource):
        ...
    """
    if request.node.get_closest_marker("enable_only_for_data_resource_type"):
        enabled_data_resource_types = request.node.get_closest_marker(
            "enable_only_for_data_resource_type"
        ).args
        assert all(
            isinstance(data_type, ServiceType) for data_type in enabled_data_resource_types
        ), "Arguments to `enable_only_for_data_resource_type()` must be of type ServiceType"

        if data_resource.type() not in enabled_data_resource_types:
            pytest.skip(
                "Skipped for data resource `%s`, since it is not of type `%s`."
                % (data_resource.name(), ",".join(enabled_data_resource_types))
            )


def pytest_sessionfinish(session, exitstatus):
    # hasattr(session.config, "workerinput") ensures
    # this only triggers after all workflow finishes.
    if not hasattr(session.config, "workerinput") and not session.config.getoption("keep_flows"):
        client = aq.Client(*get_aqueduct_config())
        delete_all_flows(client)
