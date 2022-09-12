import pytest
from utils import run_flow_test

from aqueduct import op


@op
def no_return() -> None:
    return None


@pytest.mark.publish
def test_operator_with_no_return(client):
    result = no_return()
    assert result.get() is None
    try:
        flow = run_flow_test(client, artifacts=[result], delete_flow_after=False)
        artifact_return = flow.latest().artifact("no_return artifact")
        assert artifact_return.get() is None
    finally:
        client.delete_flow(flow.id())
