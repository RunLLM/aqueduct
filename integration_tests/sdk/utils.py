import time
import uuid
from typing import Any, Dict, List, Optional, Union

from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.table_artifact import TableArtifact
from aqueduct.constants.enums import ArtifactType, ExecutionStatus, LoadUpdateMode, S3TableFormat
from aqueduct.integrations.s3_integration import S3Integration
from aqueduct.integrations.sql_integration import RelationalDBIntegration
from aqueduct.models.operators import LoadSpec, RelationalDBLoadParams, S3LoadParams
from data_objects import DataObject

import aqueduct
from aqueduct import Flow

# TODO(ENG-1738): this global dictionary is only maintained because we don't have a way
#  of deleting flows by name yet. The teardown code has the flow name, but not the flow id,
#  since that is generated in the test by `publish_flow()`. Therefore, we must register every
#  flow we publish in `publish_flow_test` in this dictionary.
flow_name_to_id: Dict[str, uuid.UUID] = {}

# Toggles whether we should test deprecated code paths. The is useful for ensuring both the new and
# old code paths continue to work when the API changes, but we want to continue to ensure backwards
# compatibility for a while.
use_deprecated_code_paths = False

# Global map tracking all the artifacts we've saved in the test suite and the path that they were saved to.
artifact_id_to_saved_identifier: Dict[str, str] = {}


def use_deprecated_code_path() -> bool:
    return use_deprecated_code_paths


def generate_new_flow_name() -> str:
    return "test_" + uuid.uuid4().hex


def generate_table_name() -> str:
    return "test_table_" + uuid.uuid4().hex[:24]


def publish_flow_test(
    client: aqueduct.Client,
    artifacts: Union[BaseArtifact, List[BaseArtifact]],
    engine: Optional[str],
    expected_statuses: Union[ExecutionStatus, List[ExecutionStatus]] = ExecutionStatus.SUCCEEDED,
    name: Optional[str] = None,
    existing_flow: Optional[Flow] = None,
    metrics: Optional[List[BaseArtifact]] = None,
    checks: Optional[List[BaseArtifact]] = None,
    schedule: str = "",
    should_block: bool = True,
) -> Flow:
    """Publishes a flow and waits for a specified number of runs with specified statuses to complete.

    Args:
        artifacts:
            These are fed directly into client.publish_flow()
        engine:
            The engine to publish against.
        expected_statuses:
            The expected outcomes of the published flow's runs. This method will not return until these
            are all satisfied or violated. It can be supplied as either a single status or a list of them.
            When supplied as a list, the statuses are interpreted in chronologically ascending order. Eg.
            [SUCCEEDED, FAILED} means I expect one successful run, followed by an unsuccessful one.

            If publishing against a flow that already exists, previous runs will be disregarded. These
            statuses are expectations about future runs of the flow.
        name:
            This is fed directly into client.publish_flow(). This is also registered with `flow_name_to_id`
            for cleanup purposes.
        existing_flow:
            If we are publishing against a flow that already exists, that flow object must be provided.
            Otherwise, `name` must be supplied.
        metrics:
        checks:
        schedule:
            These are fed directly into `publish_flow()`.
        should_block:
            When true (default), we return immediately after publishing, without waiting for the flows to complete.
            Currently, the only reason this is ever false is if you need to spin up two flows at exactly the same
            time, and then wait for both of them to complete afterwards.
    """
    assert (
        name or existing_flow and not (name and existing_flow)
    ), "Either `name` or `existing_flow` can be set, but not both."

    if existing_flow is not None:
        name = existing_flow.name()
    assert isinstance(name, str), "Flow name must be string, not %s type." % type(name)

    # Check that if a new flow name is provided, the flow really does not exist.
    if existing_flow is None:
        flow_dicts = client.list_flows()
        assert all(
            flow_dict["name"] != name for flow_dict in flow_dicts
        ), "You are publishing with a flow name that has already been published, please supply `existing_flow` instead."

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
            expected_statuses=[expected_statuses]
            if isinstance(expected_statuses, ExecutionStatus)
            else expected_statuses,
        )
    return flow


def trigger_flow_test(
    client: aqueduct.Client,
    flow: Flow,
    expected_status: Union[ExecutionStatus, List[ExecutionStatus]] = ExecutionStatus.SUCCEEDED,
    parameters: Optional[Dict[str, Any]] = None,
) -> None:
    """Triggers the given flow, and waits for the expected runs to complete with expected statuses.

    `expected_status` is interpreted the same way as in `publish_flow_test()` above.
    """
    num_prev_runs = len(flow.list_runs())
    client.trigger(flow.id(), parameters=parameters)

    wait_for_flow_runs(
        client,
        flow.id(),
        num_prev_runs=num_prev_runs,
        expected_statuses=[expected_status]
        if isinstance(expected_status, ExecutionStatus)
        else expected_status,
    )


def wait_for_flow_runs(
    client: aqueduct.Client,
    flow_id: uuid.UUID,
    expected_statuses: List[ExecutionStatus],
    num_prev_runs: int = 0,
) -> None:
    """Waits for a flow to complete len(expected_statuses) runs, with the expected statuses.

    Statuses are sorted in chronologically ascending order. `num_prev_runs` denotes the number of
    previous runs of the flow to ignore when checking new run statuses.

    NOTE: This should only ever directly be used by a test when `publish_flow_test(..., should_block=True)`.
          Otherwise, just use the publish and trigger helpers directly, instead of calling this.
    """

    def stop_condition(
        client: aqueduct.Client,
        flow_id: uuid.UUID,
        expected_statuses: List[ExecutionStatus],
        num_prev_runs: int,
    ) -> bool:
        if all(str(flow_id) != flow_dict["flow_id"] for flow_dict in client.list_flows()):
            return False

        flow = client.flow(flow_id)

        # Last run is prepended to this list. We need to reverse it in order to compare with expected statuses,
        # which is sorted in chronologically ascending order.
        flow_runs = list(reversed(flow.list_runs()))[num_prev_runs:]
        if len(flow_runs) < len(expected_statuses):
            return False

        statuses = [flow_run["status"] for flow_run in flow_runs]

        # Continue checking as long as there are still runs pending.
        if any(status == ExecutionStatus.PENDING for status in statuses):
            return False

        expect_status_strs = [status.value for status in expected_statuses]
        assert statuses == expect_status_strs, (
            "Unexpected workflow run status(es). In ascending chronological order (latest last), expected %s, got %s. "
            % (expect_status_strs, statuses)
        )

        print(
            "Workflow %s was created and ran successfully at least %s times!"
            % (flow_id, len(flow_runs))
        )
        return True

    polling(
        lambda: stop_condition(client, flow_id, expected_statuses, num_prev_runs),
        timeout=300,
        poll_threshold=1,
        timeout_comment="Timed out waiting for workflow run to complete.",
    )


def extract(
    integration,
    obj_identifier: Union[DataObject, str],
    op_name: Optional[str] = None,
    lazy: bool = False,
) -> BaseArtifact:
    """Reads the specified object in from the integration and returns it as an artifact.

    Assumption: the object is a pandas dataframe, serialized in a particular fashion in each integration.
    This serialization method should match what is done in `save()`.
    """
    if isinstance(obj_identifier, DataObject):
        obj_identifier = obj_identifier.value

    assert isinstance(obj_identifier, str)
    if isinstance(integration, RelationalDBIntegration):
        return integration.sql(query="SELECT * from %s" % obj_identifier, name=op_name, lazy=lazy)
    elif isinstance(integration, S3Integration):
        return integration.file(
            obj_identifier, ArtifactType.TABLE, "parquet", name=op_name, lazy=lazy
        )
    raise Exception("Unexpected data integration type provided in test: %s", type(integration))


def get_object_identifier_from_load_spec(load_spec: LoadSpec) -> str:
    if isinstance(load_spec.parameters, RelationalDBLoadParams):
        return load_spec.parameters.table
    elif isinstance(load_spec.parameters, S3LoadParams):
        return load_spec.parameters.filepath
    raise Exception("Unsupported load parameters %s." % type(load_spec.parameters))


def set_object_identifier_on_load_spec(load_spec: LoadSpec, new_obj_identifier: str) -> str:
    if isinstance(load_spec.parameters, RelationalDBLoadParams):
        load_spec.parameters.table = new_obj_identifier
    elif isinstance(load_spec.parameters, S3LoadParams):
        load_spec.parameters.filepath = new_obj_identifier
    else:
        raise Exception("Unsupported load parameters %s." % type(load_spec.parameters))


def save(
    integration,
    artifact: TableArtifact,
    name: Optional[str] = None,
    update_mode: Optional[LoadUpdateMode] = None,
):
    """Saves an artifact into the integration.

    If `name` is set, make sure that it is set to a globally unique value, since test cases can be run concurrently.

    Assumption: the artifact represents a pandas dataframe. Each type of integration is serialized in a particular fashion.
    It should match the deserialization method in extract().
    """
    if name is None:
        name = generate_table_name()
    if update_mode is None:
        update_mode = LoadUpdateMode.REPLACE

    if isinstance(integration, RelationalDBIntegration):
        if use_deprecated_code_path():
            artifact.save(integration.config(name, update_mode))
        else:
            integration.save(artifact, name, update_mode)

    elif isinstance(integration, S3Integration):
        assert update_mode == LoadUpdateMode.REPLACE, "S3 only supports replacement update."
        integration.save(artifact, name, S3TableFormat.PARQUET)

        # Record where the artifact was saved, so we can validate the data later, after the flow is published.
        artifact_id_to_saved_identifier[str(artifact.id())] = name
    else:
        raise Exception("Unexpected data integration type provided in test: %s", type(integration))

    # Record where the artifact was saved, so we can validate the data later, after the flow is published.
    artifact_id_to_saved_identifier[str(artifact.id())] = name


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
