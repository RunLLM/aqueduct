import time
import uuid
from typing import Dict, List, Optional

from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.enums import ExecutionStatus

import aqueduct
from aqueduct import Flow


def generate_new_flow_name() -> str:
    return "test_" + uuid.uuid4().hex


def generate_table_name() -> str:
    return "test_table_" + uuid.uuid4().hex[:24]


def publish_flow(
    client: aqueduct.Client,
    artifacts: List[BaseArtifact],
    metrics: Optional[List[BaseArtifact]] = None,
    checks: Optional[List[BaseArtifact]] = None,
    schedule: str = "",
):

    # TODO: register with context that will best-effort delete afterwards.
    pass


def run_flow_test(
    client: aqueduct.Client,
    artifacts: List[BaseArtifact],
    engine: Optional[str],
    metrics: Optional[List[BaseArtifact]] = None,
    checks: Optional[List[BaseArtifact]] = None,
    name: str = "",
    schedule: str = "",
    num_runs: int = 1,
    delete_flow_after: bool = True,
    expect_success: bool = True,
) -> Optional[Flow]:
    """
    Publishes the flow and waits until it has run at least `num_runs` times with the expected status.
    The flow is always deleted before this method returns, unless `delete_flow_after = False`.
    """
    if len(name) == 0:
        name = generate_new_flow_name()

    flow = client.publish_flow(
        name=name,
        artifacts=artifacts,
        engine=engine,
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
