import json
import textwrap
import uuid
from collections import defaultdict
from typing import DefaultDict, Dict, List, Union

from aqueduct.dag import DAG
from aqueduct.error import InvalidUserActionException, InvalidUserArgumentException

from aqueduct import globals

from .enums import ArtifactType, OperatorType
from .flow_run import FlowRun
from .logger import logger
from .operators import OperatorSpec, ParamSpec
from .responses import SavedObjectUpdate, WorkflowDagResponse, WorkflowDagResultResponse
from .utils import format_header_for_print, generate_ui_url, parse_user_supplied_id


class Flow:
    """This class is a read-only handle to flow already in the system.

    Flows can have at multiple corresponding runs, and must have at least one.
    """

    def __init__(
        self,
        flow_id: str,
        in_notebook_or_console_context: bool,
    ):
        assert flow_id is not None
        self._id = flow_id
        self._in_notebook_or_console_context = in_notebook_or_console_context

    def id(self) -> uuid.UUID:
        """Returns the id of the flow."""
        return uuid.UUID(self._id)

    def list_runs(self, limit: int = 10) -> List[Dict[str, str]]:
        """Lists the historical runs associated with this flow, sorted chronologically from most to least recent.

        Args:
            limit:
                If set, we return only a limit number of the latest runs. Defaults to 10.

        Returns:
            A list of dictionaries, each of which corresponds to a single flow run.
            Each dictionary contains essential information about the run (eg. id, status, etc.).
        """
        if not isinstance(limit, int) or limit < 0:
            raise InvalidUserArgumentException("Limit must be a positive integer.")

        resp = globals.__GLOBAL_API_CLIENT__.get_workflow(self._id)
        return [
            dag_result.to_readable_dict()
            for dag_result in list(reversed(resp.workflow_dag_results))[:limit]
        ]

    def _construct_flow_run(
        self, dag_result: WorkflowDagResultResponse, dag_resp: WorkflowDagResponse
    ) -> FlowRun:
        """Constructs a flow run from a GetWorkflowResponse."""
        dag = DAG(
            operators=dag_resp.operators,
            artifacts=dag_resp.artifacts,
            operator_by_name={op.name: op for op in dag_resp.operators.values()},
            metadata=dag_resp.metadata,
        )

        # The dags for fetched flow runs are missing their serialized functions.
        return FlowRun(
            flow_id=self._id,
            run_id=str(dag_result.id),
            in_notebook_or_console_context=self._in_notebook_or_console_context,
            dag=dag,
            created_at=dag_result.created_at,
            status=dag_result.status,
        )

    def latest(self) -> FlowRun:
        resp = globals.__GLOBAL_API_CLIENT__.get_workflow(self._id)
        if len(resp.workflow_dag_results) == 0:
            raise InvalidUserActionException("This flow has not been run yet.")

        latest_result = resp.workflow_dag_results[-1]
        latest_workflow_dag = resp.workflow_dags[latest_result.workflow_dag_id]
        return self._construct_flow_run(latest_result, latest_workflow_dag)

    def fetch(self, run_id: Union[str, uuid.UUID]) -> FlowRun:
        run_id = parse_user_supplied_id(run_id)

        resp = globals.__GLOBAL_API_CLIENT__.get_workflow(self._id)
        assert (
            len(resp.workflow_dag_results) > 0
        ), "Every flow must have at least one run attached to it."

        result = None
        for candidate_result in resp.workflow_dag_results:
            if str(candidate_result.id) == run_id:
                assert result is None, "Cannot have two runs with the same id."
                result = candidate_result

        if result is None:
            raise InvalidUserArgumentException(
                "Cannot find any run with id %s on this flow." % run_id
            )

        workflow_dag = resp.workflow_dags[result.workflow_dag_id]
        return self._construct_flow_run(result, workflow_dag)

    def list_saved_objects(self) -> DefaultDict[str, List[SavedObjectUpdate]]:
        """Get everything saved by the flow.

        Returns:
            A dictionary mapping the integration id to the list of table names/storage path.
        """
        workflow_objects = globals.__GLOBAL_API_CLIENT__.list_saved_objects(self._id).object_details
        object_mapping = defaultdict(list)
        for item in workflow_objects:
            object_mapping[item.integration_name].append(item)
        return object_mapping

    def describe(self) -> None:
        """Prints out a human-readable description of the flow."""
        resp = globals.__GLOBAL_API_CLIENT__.get_workflow(self._id)
        latest_result = resp.workflow_dag_results[-1]
        latest_workflow_dag = resp.workflow_dags[latest_result.workflow_dag_id]

        latest_metadata = latest_workflow_dag.metadata
        assert latest_metadata.schedule is not None, "A flow must have a schedule."
        assert latest_metadata.retention_policy is not None, "A flow must have a retention policy."

        url = generate_ui_url(globals.__GLOBAL_API_CLIENT__.construct_base_url(), self._id)

        print(
            textwrap.dedent(
                f"""
            {format_header_for_print(f"'{latest_metadata.name}' Flow")}
            ID: {self._id}
            Description: '{latest_metadata.description}'
            UI: {url}
            Schedule: {latest_metadata.schedule.json(exclude_none=True)}
            RetentionPolicy: {latest_metadata.retention_policy.json(exclude_none=True)}
            Runs:
            """
            )
        )
        print(json.dumps(self.list_runs(), sort_keys=False, indent=4))
