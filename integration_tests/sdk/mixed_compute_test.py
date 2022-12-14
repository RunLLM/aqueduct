import pytest
from aqueduct.constants.enums import ServiceType

from aqueduct import op
from sdk.constants import SENTIMENT_SQL_QUERY
from sdk.utils import publish_flow_test


@pytest.mark.enable_only_for_engine_type(ServiceType.K8S, ServiceType.LAMBDA)
def test_flow_with_multiple_compute_using_op_spec(client, flow_name, data_integration, engine):
    integration = client.integration(data_integration)

    sql_artifact = integration.sql(query=SENTIMENT_SQL_QUERY)

    @op
    def noop(input):
        return input

    @op(engine=engine, requirements=[])
    def noop_on_third_party(input):
        return input

    # Only `noop_on_third_party` is run on outside compute.
    output = noop_on_third_party(noop(sql_artifact))
    publish_flow_test(client, output, engine=None)
