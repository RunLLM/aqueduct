import pytest
from aqueduct.constants.enums import ServiceType

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

        if data_integration._metadata.service not in enabled_data_integration_types:
            pytest.skip(
                "Skipped for data integration `%s`, since it is not of type `%s`."
                % (data_integration._metadata.name, ",".join(enabled_data_integration_types))
            )
