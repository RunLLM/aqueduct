import time
import uuid
from typing import Dict, List, Optional, Union, Any

from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.enums import ExecutionStatus

import aqueduct
from aqueduct import Flow


# TODO(...): because we don't have a way to deleting a flow by name yet.
flow_name_to_id: Dict[str, uuid.UUID] = {}


# TODO: remove this
def generate_new_flow_name() -> str:
    return "test_" + uuid.uuid4().hex


def generate_table_name() -> str:
    return "test_table_" + uuid.uuid4().hex[:24]


def publish_flow_test(
    client: aqueduct.Client,
    artifacts: Union[BaseArtifact, List[BaseArtifact]],
    engine: str,
    expected_statuses: Union[ExecutionStatus, List[ExecutionStatus]] = ExecutionStatus.SUCCEEDED,
    name: Optional[str] = None,
    existing_flow: Optional[Flow] = None,
    metrics: Optional[List[BaseArtifact]] = None,
    checks: Optional[List[BaseArtifact]] = None,
    schedule: str = "",
    should_block: bool = True
) -> Flow:
    """
    TODO:
    `expected_status` can be supplied as either a single status or a list of them, depending on how many
    flow runs we want to wait for.
    What is flow for?
    """
    assert name or existing_flow and not (name and existing_flow), "Either `name` or `existing_flow` can be set, but not both."

    if existing_flow is not None:
        name = existing_flow.name()
    assert isinstance(name, str), "Flow name must be string, not %s type." % type(name)

    num_prev_runs = len(existing_flow.list_runs()) if existing_flow is not None else 0
    flow = client.publish_flow(
        name=name,
        artifacts=artifacts,
        metrics=metrics,
        checks=checks,
        schedule=schedule,
        engine=engine,
    )
    print("Workflow registration succeeded. Workflow ID %s. Name: %s" % (flow.id(), name))

    # Necessary so that the flow is cleaned up at the end of the test.
    flow_name_to_id[name] = flow.id()

    if should_block:
        wait_for_flow_runs(
            client,
            flow.id(),
            num_prev_runs=num_prev_runs,
            expected_statuses=[expected_statuses] if isinstance(expected_statuses, ExecutionStatus) else expected_statuses,
        )
    return flow


def trigger_flow_test(
    client: aqueduct.Client,
    flow: Flow,
    expected_status: Union[ExecutionStatus, List[ExecutionStatus]] = ExecutionStatus.SUCCEEDED,
    parameters: Optional[Dict[str, Any]] = None,
) -> None:
    num_prev_runs = len(flow.list_runs())
    client.trigger(flow.id(), parameters=parameters)

    wait_for_flow_runs(
        client,
        flow.id(),
        num_prev_runs=num_prev_runs,
        expected_statuses=[expected_status] if isinstance(expected_status, ExecutionStatus) else expected_status,
    )


# TODO: remove
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
            expected_statuses=[expect_status] * num_runs,
        )
    finally:
        if delete_flow_after:
            delete_flow(client, flow.id())
    return flow


# TODO: make this private. Or put a warning about using this.
def wait_for_flow_runs(
    client: aqueduct.Client,
    flow_id: uuid.UUID,
    expected_statuses: List[ExecutionStatus],
    num_prev_runs: int = 0,
) -> None:
    """
    Returns only when the specified flow has run at least len(expect_statuses) times.
    Each status expectation corresponds to a single flow run.

    Returns:
        The number of runs this flow has performed.
    """
    timeout = 300
    poll_threshold = 1
    begin = time.time()

    while True:
        time.sleep(poll_threshold)

        assert time.time() - begin < timeout, "Timed out waiting for workflow run to complete."

        if all(str(flow_id) != flow_dict["flow_id"] for flow_dict in client.list_flows()):
            continue

        flow = client.flow(flow_id)

        # Last run is prepended to this list. We need to reverse it in order to compare with expected statuses,
        # which is sorted in chronologically ascending order.
        flow_runs = list(reversed(flow.list_runs()))[num_prev_runs:]
        if len(flow_runs) < len(expected_statuses):
            continue

        statuses = [flow_run["status"] for flow_run in flow_runs]

        # Continue checking as long as there are still runs pending.
        if any(status == ExecutionStatus.PENDING for status in statuses):
            continue

        expect_status_strs = [status.value for status in expected_statuses]
        assert statuses == expect_status_strs, (
            "Unexpected workflow run status(es). In ascending chronological order (latest last), expected %s, got %s. "
            % (expect_status_strs, statuses)
        )

        print(
            "Workflow %s was created and ran successfully at least %s times!"
            % (flow_id, len(flow_runs))
        )
        return


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
