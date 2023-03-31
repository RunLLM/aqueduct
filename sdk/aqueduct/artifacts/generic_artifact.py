from __future__ import annotations

import json
import uuid
from typing import Any, Dict, Optional

from aqueduct.artifacts import numeric_artifact
from aqueduct.artifacts import preview as artifact_utils
from aqueduct.artifacts import system_metric
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.constants.enums import ArtifactType, ExecutionStatus
from aqueduct.error import ArtifactNeverComputedException
from aqueduct.models.dag import DAG
from aqueduct.utils.utils import format_header_for_print

from aqueduct import globals


class GenericArtifact(BaseArtifact, system_metric.SystemMetricMixin):
    """This class represents a generic artifact within the flow's DAG.

    Currently, a generic artifact can be any artifact other than table, numeric, bool, or parameter
    generated from eager execution, or an artifact of unknown type generated from lazy execution.
    """

    def __init__(
        self,
        dag: DAG,
        artifact_id: uuid.UUID,
        artifact_type: ArtifactType = ArtifactType.UNTYPED,
        content: Optional[Any] = None,
        from_flow_run: bool = False,
        execution_status: Optional[ExecutionStatus] = None,
    ):
        # Cannot initialize a generic artifact's content without also setting its type.
        if content is not None:
            assert artifact_type != ArtifactType.UNTYPED

        self._dag = dag
        self._artifact_id = artifact_id

        # This parameter indicates whether the artifact is fetched from flow-run or not.
        self._from_flow_run = from_flow_run
        self._set_content(content)
        # This is only relevant to generic artifact produced from flow_run.artifact().
        # We need this to distinguish between when an artifact's content is None versus
        # when it fails to compute successfully.
        self._execution_status = execution_status

    def get(self, parameters: Optional[Dict[str, Any]] = None) -> Any:
        """Materializes the artifact.

        Returns:
            The materialized value.

        Raises:
            InvalidRequestError:
                An error occurred because of an issue with the user's code or inputs.
            InternalServerError:
                An unexpected error occurred in the server.
        """
        self._dag.must_get_artifact(self._artifact_id)

        if self._from_flow_run:
            if self._execution_status != ExecutionStatus.SUCCEEDED:
                raise ArtifactNeverComputedException(
                    "This artifact was part of an existing flow run but was never computed successfully!",
                )
            elif parameters is not None:
                raise NotImplementedError(
                    "Parameterizing historical artifacts is not currently supported."
                )
            return self._get_content()

        content = self._get_content()
        if parameters is not None or content is None:
            previewed_artifact = artifact_utils.preview_artifact(
                self._dag, self._artifact_id, parameters
            )
            content = previewed_artifact._get_content()

            # If the artifact was previously generated lazily, materialize the contents.
            if parameters is None and self._get_content() is None:
                self._set_content(content)

        return content

    def describe(self) -> None:
        """Prints out a human-readable description of the bool artifact."""
        input_operator = self._dag.must_get_operator(with_output_artifact_id=self._artifact_id)
        readable_dict = super()._describe()
        readable_dict["Inputs"] = [
            self._dag.must_get_artifact(artf).name for artf in input_operator.inputs
        ]
        print(format_header_for_print(f"'{input_operator.name}' {self.type()} Artifact"))
        print(json.dumps(readable_dict, sort_keys=False, indent=4))

    def system_metric(
        self, metric_name: str, lazy: bool = False
    ) -> numeric_artifact.NumericArtifact:
        """Creates a system metric that represents the given system information from the previous @op that ran on the table.

        Args:
            metric_name:
                Name of system metric to retrieve for the table.
                Valid metrics are:
                    runtime: runtime of previous @op func in seconds
                    max_memory: maximum memory usage of previous @op func in Mb

        Returns:
            A numeric artifact that represents the requested system metric
        """
        if globals.__GLOBAL_CONFIG__.lazy:
            lazy = True

        return self._system_metric_helper(self._dag, self._artifact_id, metric_name, lazy)
