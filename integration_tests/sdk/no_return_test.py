import pytest
from aqueduct import op
from utils import run_flow_test


@op
def no_return() -> None:
    return None


@pytest.mark.publish
def test_operator_with_no_return(client):
    result = no_return()
    assert(result.get() is None)
    flow = run_flow_test(client, artifacts=[result], delete_flow_after=False)

    try:
        artifact_return = flow.latest().artifact("no_return artifact")
        assert(artifact_return.get() is None)
    finally:
        client.delete_flow(flow.id())
