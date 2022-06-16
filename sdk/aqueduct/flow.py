import json
import textwrap
import uuid
from typing import Dict, List, Union

from aqueduct.api_client import APIClient
from aqueduct.dag import DAG
from aqueduct.error import InvalidUserArgumentException
from .enums import ArtifactType

from .flow_run import FlowRun
from .responses import WorkflowDagResponse
from .utils import parse_user_supplied_id, format_header_for_print


class Flow:
    """This class is a read-only handle to a workflow that in the system.

    A flow can have multiple runs within it.
    """

    def __init__(
        self,
        api_client: APIClient,
        flow_id: str,
        in_notebook_or_console_context: bool,
    ):
        assert flow_id is not None
        self._api_client = api_client
        self._id = flow_id
        self._in_notebook_or_console_context = in_notebook_or_console_context

    def id(self) -> uuid.UUID:
        """Returns the id of the flow."""
        return uuid.UUID(self._id)

    def list_runs(self) -> List[Dict[str, str]]:
        # TODO: docstring (note that this is in reverse order)
        resp = self._api_client.get_workflow(self._id)
        return [
            dag_result.to_readable_dict()
            for dag_result in reversed(resp.workflow_dag_results)
        ]

    def _reconstruct_dag(self, dag_result_id: uuid.UUID, dag_resp: WorkflowDagResponse) -> DAG:
        """TODO: docstring"""
        dag = DAG(
            operators=dag_resp.operators,
            artifacts=dag_resp.artifacts,
            operator_by_name={
                op.name: op for op in dag_resp.operators.values()
            },
            metadata=dag_resp.metadata,
        )

        # Update all parameter artifacts to the
        param_artifacts = dag.list_artifacts(filter_to=[ArtifactType.PARAM])
        for param_artifact in param_artifacts:
            assert param_artifact.spec.jsonable is not None

            param_val = self._api_client.get_artifact_result_data(
                str(dag_result_id),
                str(param_artifact.id),
            )

            # Skip parameter update if the parameter was never computed.
            # TODO(this is bug):
            if len(param_val) == 0:
                continue

            param_op = dag.must_get_operator(with_output_artifact_id=param_artifact.id)
            param_op.spec.param.val = param_val
            dag.update_operator(param_op)
        return dag

    def latest(self) -> FlowRun:
        resp = self._api_client.get_workflow(self._id)
        assert len(resp.workflow_dag_results) > 0, "Every flow must have at least one run attached to it."

        latest_result = resp.workflow_dag_results[-1]
        latest_workflow_dag = resp.workflow_dags[latest_result.workflow_dag_id]
        return FlowRun(
            api_client=self._api_client,
            run_id=str(latest_result.id),
            in_notebook_or_console_context=self._in_notebook_or_console_context,
            dag=self._reconstruct_dag(latest_workflow_dag),
            created_at=latest_result.created_at,
            status=latest_result.status,
        )

    def fetch(self, run_id: Union[str, uuid.UUID]) -> FlowRun:
        run_id = parse_user_supplied_id(run_id)

        resp = self._api_client.get_workflow(self._id)
        assert len(resp.workflow_dag_results) > 0, "Every flow must have at least one run attached to it."

        result = None
        for candidate_result in resp.workflow_dag_results:
            if str(candidate_result.id) == run_id:
                assert result is None, "Cannot have two runs with the same id."
                result = candidate_result

        if result is None:
            raise InvalidUserArgumentException("Cannot find any run with id %s on this flow." % run_id)

        workflow_dag = resp.workflow_dags[result.workflow_dag_id]
        return FlowRun(
            api_client=self._api_client,
            run_id=str(result.id),
            in_notebook_or_console_context=self._in_notebook_or_console_context,
            dag=self._reconstruct_dag(workflow_dag),
            created_at=result.created_at,
            status=result.status,
        )

    def describe(self) -> None:
        """Prints out a human-readable description of the flow."""
        resp = self._api_client.get_workflow(self._id)
        latest_result = resp.workflow_dag_results[-1]
        latest_workflow_dag = resp.workflow_dags[latest_result.workflow_dag_id]

        latest_metadata = latest_workflow_dag.metadata
        print(textwrap.dedent(
            f"""
            {format_header_for_print(f"'{latest_metadata.name}' Flow")}
            ID: {self._id}
            Description: '{latest_metadata.description}'
            Schedule: {latest_metadata.schedule.json(exclude_none=True)}
            RetentionPolicy: {latest_metadata.retention_policy.json(exclude_none=True)}
            """
        ))
        print(json.dumps(self.list_runs(), sort_keys=False, indent=4))
