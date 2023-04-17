import platform

import pytest

from aqueduct import check, global_config, op
from aqueduct.error import AqueductError

from ..shared.data_objects import DataObject
from ..shared.flow_helpers import publish_flow_test
from .extract import extract


@pytest.mark.enable_only_for_external_compute()
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
    flow = publish_flow_test(client, output, name=flow_name(), engine=None)

    flow_run = flow.latest()
    assert flow_run.artifact("noop artifact").get().equals(table_artifact.get())
    assert flow_run.artifact("noop_on_third_party artifact").get().equals(table_artifact.get())


@pytest.mark.skip_for_spark_engines(reason="Cannot switch between Spark and Aqueduct engines.")
@pytest.mark.enable_only_for_external_compute()
def test_global_config_engine_switching(client, engine):
    """Test that we can freely switch between Aqueduct to External compute and back again."""

    @check(severity="error", requirements=[])
    def must_be_local(node):
        return node == platform.node()

    @check(severity="error", requirements=[])
    def must_be_external(node):
        return node != platform.node()

    # This should match the executor's info if executed locally.
    current_machine_node = platform.node()

    # First, verify that we are currently running locally, not externally.
    _ = must_be_local(current_machine_node)
    with pytest.raises(AqueductError, match="The check did not pass \(returned False\)"):
        _ = must_be_external(current_machine_node)

    # Next, change our compute to the external engine.
    global_config({"engine": engine})
    _ = must_be_external(current_machine_node)

    # Finally, switch back to our local machine.
    global_config({"engine": "Aqueduct"})
    _ = must_be_local(current_machine_node)
