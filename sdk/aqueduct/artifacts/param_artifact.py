import json
import textwrap
import uuid
from typing import Any, Dict, Optional

from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.dag import DAG
from aqueduct.error import ArtifactNeverComputedException, InvalidUserArgumentException
from aqueduct.utils import format_header_for_print


class ParamArtifact(BaseArtifact):
    def __init__(
        self,
        dag: DAG,
        artifact_id: uuid.UUID,
    ):
        """The APIClient is only included because decorated functions operators acting on this parameter
        will need a handle to an API client."""

        self._dag = dag
        self._artifact_id = artifact_id
        self._from_flow_run = False

        # Set the content of this parameter from the SDK dag's operator spec.
        param_op = self._dag.must_get_operator(with_output_artifact_id=self._artifact_id)
        assert param_op.spec.param is not None, "Artifact is not a parameter."
        self._set_content(json.loads(param_op.spec.param.val))

    def get(self, parameters: Optional[Dict[str, Any]] = None) -> Any:
        if parameters is not None:
            raise InvalidUserArgumentException(
                "Parameters cannot be supplied to parameter artifacts. They should only be provided"
                "for artifact's that depend on parameters."
            )

        if self._from_flow_run and self._get_content() is None:
            raise ArtifactNeverComputedException(
                "This artifact was part of an existing flow run but was never computed successfully!",
            )

        return self._get_content()

    def describe(self) -> None:
        print(
            textwrap.dedent(
                f"""
            {format_header_for_print(f"'{self.name()}' Parameter")}
            Value: {self.get()}
            """
            )
        )
