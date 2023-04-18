import textwrap
import uuid
from datetime import datetime
from typing import Optional

from aqueduct.artifacts import (
    base_artifact,
    bool_artifact,
    generic_artifact,
    numeric_artifact,
    table_artifact,
)
from aqueduct.constants.enums import ArtifactType, ExecutionStatus, OperatorType
from aqueduct.error import InternalAqueductError
from aqueduct.models.dag import DAG
from aqueduct.models.utils import TIME_FORMAT, human_readable_timestamp
from aqueduct.utils.utils import (
    format_header_for_print,
    generate_ui_url,
    indent_multiline_string,
    print_logs,
)

from aqueduct import globals


class FlowRun:
    """This class is a read-only handle corresponding to a single workflow run in the system."""

    def __init__(
        self,
        flow_id: str,
        run_id: str,
        in_notebook_or_console_context: bool,
    ):
        assert run_id is not None
        self._flow_id = flow_id
        self._id = run_id
        self._in_notebook_or_console_context = in_notebook_or_console_context

        dag_result_resp = globals.__GLOBAL_API_CLIENT__.get_workflow_dag_result(
            self._flow_id,
            self._id,
        )

        # Note that the operators for fetched flow runs are missing their serialized functions.
        self._dag = DAG(
            operators={
                str(id): elem.to_operator() for id, elem in dag_result_resp.operators.items()
            },
            artifacts={
                str(id): elem.to_artifact() for id, elem in dag_result_resp.artifacts.items()
            },
            metadata=dag_result_resp.metadata(),
        )

        self._dag_result_resp = dag_result_resp

    def id(self) -> uuid.UUID:
        """Returns the id for this flow run."""
        return uuid.UUID(self._id)

    def status(self) -> ExecutionStatus:
        """Returns the status of the flow run."""
        return self._dag_result_resp.result.exec_state.status

    def _created_at(self) -> datetime:
        """Returns the datetime at which the flow run was created."""
        return self._dag_result_resp.dag_created_at

    def created_at(self) -> float:
        """Returns the unix timestamp at which the flow run was created."""
        return self._created_at().timestamp()

    def describe(self) -> None:
        """Prints out a human-readable description of the flow run."""
        url = generate_ui_url(
            globals.__GLOBAL_API_CLIENT__.construct_base_url(),
            self._flow_id,
            self._id,
        )

        print(
            textwrap.dedent(
                f"""
            {format_header_for_print(f"'{self._dag.metadata.name}' Run")}
            ID: {self._id}
            Created At (UTC): {self._created_at().strftime(TIME_FORMAT)}
            Status: {str(self.status())}
            UI: {url}
            """
            )
        )

        param_operators = self._dag.list_operators(filter_to=[OperatorType.PARAM])
        if len(param_operators) > 0:
            print(format_header_for_print("Parameters"))
            for param_op in param_operators:
                (
                    param_content,
                    execution_status,
                ) = globals.__GLOBAL_API_CLIENT__.get_artifact_result_data(
                    self._id, str(param_op.outputs[0])
                )

                if execution_status != ExecutionStatus.SUCCEEDED:
                    param_content = "Parameter not successfully initialized."

                print("* " + param_op.name + ": " + str(param_content))

        if self.status() == ExecutionStatus.FAILED:
            print(format_header_for_print("Failures"))

            # Print out any workflow-level errors.
            dag_exec_state = self._dag_result_resp.result.exec_state
            if dag_exec_state.error is not None:
                print("Workflow-level error: ")
                print(
                    indent_multiline_string(
                        "%s\n\n%s" % (dag_exec_state.error.tip, dag_exec_state.error.context)
                    )
                )

                workflow_logs = dag_exec_state.user_logs
                if workflow_logs is not None and not workflow_logs.is_empty():
                    print("Workflow logs: ")
                    print_logs(workflow_logs)

            # Lastly, print out any operator-level errors.
            for op_result in self._dag_result_resp.operators.values():
                if op_result.result is None:
                    continue

                if op_result.result.exec_state.status == ExecutionStatus.FAILED:
                    if op_result.result.exec_state.error is not None:
                        print(f"Operator '{op_result.name}' failed: ")
                        print(
                            indent_multiline_string(
                                "%s\n%s"
                                % (
                                    op_result.result.exec_state.error.tip,
                                    op_result.result.exec_state.error.context,
                                )
                            )
                        )

                    logs = op_result.result.exec_state.user_logs
                    if logs is not None and not logs.is_empty():
                        print("Operator logs: ")
                        print_logs(logs)

    def artifact(self, name: str) -> Optional[base_artifact.BaseArtifact]:
        """Gets the Artifact from the flow run based on the name of the artifact.

        Args:
            name:
                the name of the artifact.

        Returns:
            A input artifact obtained from the dag attached to the flow run.
            If the artifact does not exist, return None.
        """
        flow_run_dag = self._dag
        artifact_from_dag = flow_run_dag.get_artifact_by_name(name)

        if artifact_from_dag is None:
            return None

        content, execution_status = globals.__GLOBAL_API_CLIENT__.get_artifact_result_data(
            self._id, str(artifact_from_dag.id)
        )

        if not isinstance(artifact_from_dag.type, ArtifactType):
            raise InternalAqueductError("The artifact's type can not be recognized.")

        if artifact_from_dag.type is ArtifactType.TABLE:
            return table_artifact.TableArtifact(
                self._dag,
                artifact_from_dag.id,
                content=content,
                from_flow_run=True,
            )
        elif artifact_from_dag.type is ArtifactType.NUMERIC:
            return numeric_artifact.NumericArtifact(
                self._dag,
                artifact_from_dag.id,
                content=content,
                from_flow_run=True,
            )
        elif artifact_from_dag.type is ArtifactType.BOOL:
            return bool_artifact.BoolArtifact(
                self._dag,
                artifact_from_dag.id,
                content=content,
                from_flow_run=True,
            )
        else:
            return generic_artifact.GenericArtifact(
                self._dag,
                artifact_from_dag.id,
                artifact_from_dag.type,
                content=content,
                from_flow_run=True,
                execution_status=execution_status,
            )
