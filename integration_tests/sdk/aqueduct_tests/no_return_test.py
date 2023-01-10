from aqueduct import op

from ..shared.utils import publish_flow_test


@op
def no_return() -> None:
    return None


def test_operator_with_no_return(client, flow_name, engine):
    result = no_return()
    assert result.get() is None

    flow = publish_flow_test(
        client,
        result,
        name=flow_name(),
        engine=engine,
    )
    artifact_return = flow.latest().artifact("no_return artifact")
    assert artifact_return.get() is None
