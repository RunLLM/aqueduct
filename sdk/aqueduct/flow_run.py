import textwrap
import uuid
from textwrap import wrap
from typing import Any, Dict, List, Mapping, Optional, Union

import plotly.graph_objects as go
from aqueduct.artifacts import (
    base_artifact,
    bool_artifact,
    generic_artifact,
    numeric_artifact,
    param_artifact,
    table_artifact,
)
from aqueduct.artifacts.metadata import ArtifactMetadata
from aqueduct.dag import DAG
from aqueduct.enums import ArtifactType, DisplayNodeType, ExecutionStatus, OperatorType
from aqueduct.error import InternalAqueductError
from aqueduct.operators import Operator
from aqueduct.utils import format_header_for_print, generate_ui_url, human_readable_timestamp

from aqueduct import api_client


class FlowRun:
    """This class is a read-only handle corresponding to a single workflow run in the system."""

    def __init__(
        self,
        flow_id: str,
        run_id: str,
        in_notebook_or_console_context: bool,
        dag: DAG,
        created_at: int,
        status: ExecutionStatus,
    ):
        assert run_id is not None
        self._flow_id = flow_id
        self._id = run_id
        self._in_notebook_or_console_context = in_notebook_or_console_context
        self._dag = dag
        self._created_at = created_at
        self._status = status

    def id(self) -> uuid.UUID:
        """Returns the id for this flow run."""
        return uuid.UUID(self._id)

    def status(self) -> ExecutionStatus:
        """Returns the status of the flow run."""
        return self._status

    def describe(self) -> None:
        """Prints out a human-readable description of the flow run."""

        url = generate_ui_url(
            api_client.__GLOBAL_API_CLIENT__.construct_base_url(),
            self._flow_id,
            self._id,
        )

        print(
            textwrap.dedent(
                f"""
            {format_header_for_print(f"'{self._dag.metadata.name}' Run")}
            ID: {self._id}
            Created At (UTC): {human_readable_timestamp(self._created_at)}
            Status: {str(self._status)}
            UI: {url}
            """
            )
        )

        param_artifacts = self._dag.list_artifacts(filter_to=[ArtifactType.PARAM])
        print(format_header_for_print("Parameters "))
        for param_artifact in param_artifacts:
            param_op = self._dag.must_get_operator(with_output_artifact_id=param_artifact.id)
            assert param_op.spec.param is not None, "Artifact is not a parameter."
            print("* " + param_op.name + ": " + param_op.spec.param.val)

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
        artifact_from_dag = flow_run_dag.get_artifacts_by_name(name)

        if artifact_from_dag is None:
            return None

        if not isinstance(artifact_from_dag.type, ArtifactType):
            raise InternalAqueductError("The artifact's type can not be recognized.")

        if artifact_from_dag.type is ArtifactType.TABLE:
            return table_artifact.TableArtifact(self._dag, artifact_from_dag.id, from_flow_run=True)
        elif artifact_from_dag.type is ArtifactType.NUMERIC:
            return numeric_artifact.NumericArtifact(
                self._dag, artifact_from_dag.id, from_flow_run=True
            )
        elif artifact_from_dag.type is ArtifactType.BOOL:
            return bool_artifact.BoolArtifact(self._dag, artifact_from_dag.id, from_flow_run=True)
        elif artifact_from_dag.type is ArtifactType.PARAM:
            return param_artifact.ParamArtifact(self._dag, artifact_from_dag.id, from_flow_run=True)
        else:
            return generic_artifact.GenericArtifact(
                self._dag, artifact_from_dag.id, artifact_from_dag.type, from_flow_run=True
            )
