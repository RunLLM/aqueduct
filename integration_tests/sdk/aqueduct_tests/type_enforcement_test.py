from typing import Union

import pytest
from aqueduct.constants.enums import ExecutionStatus
from aqueduct.error import AqueductError

from aqueduct import op

from ..shared.flow_helpers import publish_flow_test, trigger_flow_test


@op
def output_different_types(should_return_num: bool) -> Union[str, int]:
    if should_return_num:
        return 123
    return "not a number"


def test_flow_fails_on_unexpected_type_output(client, flow_name, engine):
    type_toggle = client.create_param("output_type_toggle", True)
    output = output_different_types(type_toggle)

    flow = publish_flow_test(client, name=flow_name(), artifacts=output, engine=engine)
    trigger_flow_test(
        client,
        flow,
        parameters={"output_type_toggle": False},
        expected_status=ExecutionStatus.FAILED,
    )


def test_flow_fails_on_unexpected_type_output_for_lazy(client, flow_name, engine):
    type_toggle = client.create_param("output_type_toggle", True)
    output = output_different_types.lazy(type_toggle)

    # The flow will first be lazily executed, and the new type information
    # will be persisted to the database.
    flow = publish_flow_test(client, name=flow_name(), artifacts=[output], engine=engine)

    # Because we are violating our inferred types, this will fail!
    trigger_flow_test(
        client,
        flow,
        parameters={"output_type_toggle": False},
        expected_status=ExecutionStatus.FAILED,
    )


@pytest.mark.skip_for_global_lazy_execution(reason="This test requires global eager execution.")
def test_preview_artifact_backfilled_with_wrong_type(client):
    """An error should be thrown if an upstream operator is previewed with the wrong type."""
    type_toggle = client.create_param("output_type_toggle", True)

    # Sets the type of this artifact eagerly due to preview.
    output = output_different_types(type_toggle)

    @op
    def noop(data):
        return data

    # Lazily execute the downstream operator with a custom parameter.
    # We execute lazily so that the terminal node can expect any output and therefore won't error.
    # We want to induce the upstream type backfill to error.
    noop_output = noop.lazy(output)

    # Fails because the upstream operator should have a type mismatch.
    with pytest.raises(AqueductError, match="Operator `output_different_types` failed!"):
        noop_output.get(parameters={"output_type_toggle": False})


def test_list_and_tuple_types_are_different(client):
    """
    Because we json-serialize both of these into the same bytes representation,
    make sure type fidelity is actually preserved for each.
    """

    @op
    def return_list():
        return [1, 2, 3]

    @op
    def return_tuple():
        return (1, 2, 3)

    list_output = return_list()
    assert isinstance(list_output.get(), list)
    assert list_output.get() == [1, 2, 3]

    tuple_output = return_tuple()
    assert isinstance(tuple_output.get(), tuple)
    assert tuple_output.get() == (1, 2, 3)
