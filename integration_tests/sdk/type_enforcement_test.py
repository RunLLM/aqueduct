from typing import Union

from aqueduct.enums import ExecutionStatus
from utils import run_flow_test, wait_for_flow_runs

from aqueduct import op


@op
def output_different_types(should_return_num: bool) -> Union[str, int]:
    if should_return_num:
        return 123
    return "not a number"


def test_flow_fails_on_unexpected_type_output(client):
    type_toggle = client.create_param("output_type_toggle", True)
    output = output_different_types(type_toggle)

    flow = run_flow_test(client, artifacts=[output], delete_flow_after=False)

    try:
        client.trigger(flow.id(), parameters={"output_type_toggle": False})
        wait_for_flow_runs(
            client,
            flow.id(),
            expect_statuses=[ExecutionStatus.SUCCEEDED, ExecutionStatus.FAILED],
        )
    finally:
        client.delete_flow(flow.id())
