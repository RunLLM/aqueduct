import pytest
from aqueduct.constants.enums import ServiceType

from aqueduct import op

from ..shared.data_objects import DataObject
from ..shared.flow_helpers import publish_flow_test
from .extract import extract


@pytest.mark.enable_only_for_engine_type(ServiceType.K8S, ServiceType.LAMBDA)
def test_flow_with_multiple_compute_using_op_spec(client, flow_name, data_integration, engine):
    """Runs a workflow both Aqueduct and a third-party compute engine."""
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    @op
    def noop(input):
        return input

    @op(engine=engine, requirements=[])
    def noop_on_third_party(input):
        return input

    # Only `noop_on_third_party` is run on outside compute.
    output = noop_on_third_party(noop(table_artifact))
    publish_flow_test(client, output, name=flow_name(), engine=None)
