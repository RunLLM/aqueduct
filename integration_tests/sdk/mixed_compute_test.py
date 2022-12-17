import pytest
from aqueduct.constants.enums import ServiceType
from constants import SENTIMENT_SQL_QUERY
from utils import publish_flow_test

from aqueduct import op


@pytest.mark.enable_only_for_engine_type(ServiceType.K8S, ServiceType.LAMBDA)
def test_flow_with_multiple_compute_using_op_spec(client, flow_name, data_integration, engine):
    sql_artifact = data_integration.sql(query=SENTIMENT_SQL_QUERY)

    @op
    def noop(input):
        return input

    @op(engine=engine, requirements=[])
    def noop_on_third_party(input):
        return input

    # Only `noop_on_third_party` is run on outside compute.
    output = noop_on_third_party(noop(sql_artifact))
    publish_flow_test(client, output, engine=None)
