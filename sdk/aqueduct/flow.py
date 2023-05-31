import json
import textwrap
import uuid
from collections import defaultdict
from typing import DefaultDict, Dict, List, Optional, Union

from aqueduct.error import InvalidUserArgumentException
from aqueduct.flow_run import FlowRun
from aqueduct.models.response_models import (
    GetWorkflowV1Response,
    SavedObjectUpdate,
    WorkflowDagResponse,
)
from aqueduct.utils.utils import format_header_for_print, generate_ui_url, parse_user_supplied_id

from aqueduct import globals


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

    def _get_workflow_resp(self) -> GetWorkflowV1Response:
        resp = globals.__GLOBAL_API_CLIENT__.get_workflow(self._id)
        return resp

    def name(self) -> str:
        """Returns the latest name of the flow."""
        latest_workflow_dag = self._get_latest_dag_resp()
        assert latest_workflow_dag.metadata.name is not None
        return latest_workflow_dag.metadata.name

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

    def _get_latest_dag_resp(self) -> WorkflowDagResponse:
        resp = self._get_workflow_resp()
        if not resp.workflow_dag_results:
            assert bool(resp.workflow_dags)
            return list(resp.workflow_dags.values())[0]

        latest_result = resp.workflow_dag_results[-1]
        return resp.workflow_dags[latest_result.workflow_dag_id]

    def latest(self) -> Optional[FlowRun]:
        resp = self._get_workflow_resp()
        if not resp.workflow_dag_results:
            return None

        latest_result = resp.workflow_dag_results[-1]
        return FlowRun(
            flow_id=self._id,
            run_id=str(latest_result.id),
            in_notebook_or_console_context=self._in_notebook_or_console_context,
        )

    def fetch(self, run_id: Union[str, uuid.UUID]) -> FlowRun:
        run_id = parse_user_supplied_id(run_id)

        resp = self._get_workflow_resp()
        found = False
        for candidate_result in resp.workflow_dag_results:
            if str(candidate_result.id) == run_id:
                assert not found, "Cannot have two runs with the same id."
                found = True

        if not found:
            raise InvalidUserArgumentException(
                "Cannot find any run with id %s on this flow." % run_id
            )

        return FlowRun(
            flow_id=self._id,
            run_id=str(run_id),
            in_notebook_or_console_context=self._in_notebook_or_console_context,
        )

    def list_saved_objects(self) -> DefaultDict[str, List[SavedObjectUpdate]]:
        """Get everything saved by the flow.

        Returns:
            A dictionary mapping the resource id to the list of table names/storage path.
        """
        object_mapping: DefaultDict[str, List[SavedObjectUpdate]] = defaultdict(list)

        workflow_objects = globals.__GLOBAL_API_CLIENT__.list_saved_objects(self._id).object_details
        if workflow_objects is None:
            return object_mapping  # Empty map

        object_mapping = defaultdict(list)
        for item in workflow_objects:
            object_mapping[item.resource_name].append(item)
        return object_mapping

    def describe(self) -> None:
        """Prints out a human-readable description of the flow."""
        latest_workflow_dag = self._get_latest_dag_resp()

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
            """
            )
        )

        runs = self.list_runs()
        if runs:
            print("Runs:")
            print(json.dumps(runs, sort_keys=False, indent=4))
