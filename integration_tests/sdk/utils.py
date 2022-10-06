import time
import uuid
from typing import Dict, List, Optional, Union

from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.bool_artifact import BoolArtifact
from aqueduct.artifacts.numeric_artifact import NumericArtifact
from aqueduct.artifacts.table_artifact import TableArtifact
from aqueduct.enums import ExecutionStatus
from pandas import DataFrame

# Should be set before each test runs.
from test_functions.sentiment.model import sentiment_model, sentiment_model_multiple_input
from test_functions.simple.model import dummy_sentiment_model, dummy_sentiment_model_multiple_input

import aqueduct
from aqueduct import Flow, api_client

flags: Dict[str, bool] = {}
integration_name: Optional[str] = None


def get_integration_name() -> str:
    assert integration_name is not None
    return integration_name


def should_publish_flows() -> bool:
    assert "publish" in flags
    return flags["publish"]


def should_run_complex_models() -> bool:
    assert "complex_models" in flags
    return flags["complex_models"]


def generate_new_flow_name() -> str:
    return "test_" + uuid.uuid4().hex


def generate_table_name() -> str:
    return "test_table_" + uuid.uuid4().hex[:24]


def run_sentiment_model(artifact: TableArtifact) -> TableArtifact:
    """
    Calls the full sentiment model if --complex_models flag is set. Otherwise, will use simple model,
    which appends the same column with a dummy value, but is much faster.
    """
    if should_run_complex_models():
        return sentiment_model(artifact)
    else:
        return dummy_sentiment_model(artifact)


def run_sentiment_model_local(artifact: TableArtifact) -> DataFrame:
    """
    Run sentiment model locally using .local() method. Calls the full sentiment model
    local method if --complex_models flag is set. Otherwise, will use simple model,which
    appends the same column with a dummy value.
    """
    if should_run_complex_models():
        return sentiment_model.local(artifact)
    else:
        return dummy_sentiment_model.local(artifact)


def run_sentiment_model_multiple_input(
    artifact1: TableArtifact, artifact2: TableArtifact
) -> TableArtifact:
    """
    Same test setup as `run_sentiment_model`.
    """
    if should_run_complex_models():
        return sentiment_model_multiple_input(artifact1, artifact2)
    else:
        return dummy_sentiment_model_multiple_input(artifact1, artifact2)


def run_sentiment_model_local_multiple_input(
    artifact1: TableArtifact, artifact2: TableArtifact
) -> DataFrame:
    """
    Same test setup as `run_sentiment_model_local` but takes in two artifacts.
    """
    if should_run_complex_models():
        return sentiment_model_multiple_input.local(artifact1, artifact2)
    else:
        return dummy_sentiment_model_multiple_input.local(artifact1, artifact2)


def run_flow_test(
    client: aqueduct.Client,
    artifacts: List[BaseArtifact],
    metrics: Optional[List[BaseArtifact]] = None,
    checks: Optional[List[BaseArtifact]] = None,
    name: str = "",
    schedule: str = "",
    num_runs: int = 1,
    delete_flow_after: bool = True,
    expect_success: bool = True,
) -> Optional[Flow]:
    """
    Actually publishes the flow if tests are run with --publish flag. This flow can be deleted
    within this method if `delete_flow_after = True`.

    If --publish is not supplied, we will instead realize all the artifacts with .get().
    The --publish case only returns when the specified flow has run successfully at least `num_runs` times.
    """
    if not should_publish_flows():
        for artifact in artifacts:
            _ = artifact.get()
        return None

    if len(name) == 0:
        name = generate_new_flow_name()

    flow = client.publish_flow(
        name=name,
        artifacts=artifacts,
        metrics=metrics,
        checks=checks,
        schedule=schedule,
    )
    print("Workflow registration succeeded. Workflow ID %s. Name: %s" % (flow.id(), name))

    try:
        expect_status = ExecutionStatus.SUCCEEDED if expect_success else ExecutionStatus.FAILED
        wait_for_flow_runs(
            client,
            flow.id(),
            expect_statuses=[expect_status] * num_runs,
        )
    finally:
        if delete_flow_after:
            delete_flow(client, flow.id())
    return flow


def wait_for_flow_runs(
    client: aqueduct.Client,
    flow_id: uuid.UUID,
    expect_statuses: List[ExecutionStatus],
) -> int:
    """
    Returns only when the specified flow has run at least len(expect_statuses) times.
    Each status expectation corresponds to a single flow run.

    Returns:
        The number of runs this flow has performed.
    """
    timeout = 500
    poll_threshold = 1
    begin = time.time()

    while True:
        time.sleep(poll_threshold)

        assert time.time() - begin < timeout, "Timed out waiting for workflow run to complete."

        if all(str(flow_id) != flow_dict["flow_id"] for flow_dict in client.list_flows()):
            continue

        # A flow has been successfully published if it makes at least one workflow run, and
        # all its workflow runs have executed successfully.
        flow = client.flow(flow_id)
        flow_runs = flow.list_runs()
        if len(flow_runs) == 0:
            continue

        statuses = [flow_run["status"] for flow_run in flow_runs]

        # Continue checking as long as there are still runs pending.
        if any(status == ExecutionStatus.PENDING for status in statuses):
            continue

        if len(flow_runs) < len(expect_statuses):
            continue

        # Need to reverse one of the lists for comparison, because the last run is always prepended in the backend response.
        expect_status_strs = [status.value for status in reversed(expect_statuses)]
        assert statuses == expect_status_strs, (
            "Unexpected workflow run status(es). In reverse chronological order, expected %s, got %s. "
            % (expect_status_strs, statuses)
        )

        print(
            "Workflow %s was created and ran successfully at least %s times!"
            % (flow_id, len(flow_runs))
        )
        return len(flow_runs)
    return -1


def delete_flow(client: aqueduct.Client, workflow_id: uuid.UUID) -> None:
    try:
        client.delete_flow(str(workflow_id))
    except Exception as e:
        print("Error deleting workflow %s with exception: %s" % (workflow_id, e))
    else:
        print("Successfully deleted workflow %s" % (workflow_id))


def check_flow_doesnt_exist(client, flow_id):
    def stop_condition(client, flow_id):
        try:
            client.flow(flow_id)
            return False
        except:
            return True

    polling(
        lambda: stop_condition(client, flow_id),
        timeout=60,
        poll_threshold=5,
        timeout_comment="Timed out checking flow doens't exist.",
    )


def check_table_doesnt_exist(integration, table):
    def stop_condition(integration, table):
        try:
            integration.sql(f"SELECT * FROM {table}").get()
            return False
        except:
            return True

    polling(
        lambda: stop_condition(integration, table),
        timeout=60,
        poll_threshold=5,
        timeout_comment="Timed out checking table doesn't exist.",
    )


def check_table_exists(integration, table):
    def stop_condition(integration, table):
        try:
            integration.sql(f"SELECT * FROM {table}").get()
            return True
        except:
            return False

    polling(
        lambda: stop_condition(integration, table),
        timeout=60,
        poll_threshold=5,
        timeout_comment="Timed out checking table doesn't exist.",
    )


def polling(
    stop_condition_fn,
    timeout=60,
    poll_threshold=5,
    timeout_comment="Timed out waiting for workflow run to complete.",
):
    begin = time.time()

    while True:
        assert time.time() - begin < timeout, timeout_comment

        if stop_condition_fn():
            break
        else:
            time.sleep(poll_threshold)
