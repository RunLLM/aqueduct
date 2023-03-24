from __future__ import annotations

import json
import uuid
from typing import Any, Dict, Optional, Union

from aqueduct.artifacts import preview as artifact_utils
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts._create import create_metric_or_check_artifact
from aqueduct.constants.enums import ArtifactType, ExecutionMode, ExecutionStatus
from aqueduct.error import ArtifactNeverComputedException
from aqueduct.models.dag import DAG
from aqueduct.utils.utils import format_header_for_print, generate_uuid
from aqueduct.artifacts import bool_artifact, numeric_artifact
from aqueduct.constants.metrics import SYSTEM_METRICS_INFO
from aqueduct.models.operators import Operator, OperatorSpec, SystemMetricSpec
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.utils.naming import default_artifact_name_from_op_name


class GenericArtifact(BaseArtifact):
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
                name of system metric to retrieve for the table.
                valid metrics are:
                    runtime: runtime of previous @op func in seconds
                    max_memory: maximum memory usage of previous @op func in Mb

        Returns:
            A numeric artifact that represents the requested system metric
        """
        if globals.__GLOBAL_CONFIG__.lazy:
            lazy = True
        execution_mode = ExecutionMode.EAGER if not lazy else ExecutionMode.LAZY

        operator = self._dag.must_get_operator(with_output_artifact_id=self._artifact_id)
        system_metric_description, system_metric_unit = SYSTEM_METRICS_INFO[metric_name]
        system_metric_name = "%s %s(%s) metric" % (operator.name, metric_name, system_metric_unit)
        op_spec = OperatorSpec(system_metric=SystemMetricSpec(metric_name=metric_name))
        new_artifact = self._apply_operator_to_table(
            op_spec,
            system_metric_name,
            system_metric_description,
            output_artifact_type_hint=ArtifactType.NUMERIC,
            execution_mode=execution_mode,
        )

        assert isinstance(new_artifact, numeric_artifact.NumericArtifact)

        return new_artifact

    def _apply_operator_to_table(
        self,
        op_spec: OperatorSpec,
        op_name: str,
        op_description: str,
        output_artifact_type_hint: ArtifactType,
        execution_mode: ExecutionMode = ExecutionMode.EAGER,
    ) -> Union[numeric_artifact.NumericArtifact, bool_artifact.BoolArtifact]:
        assert (
            output_artifact_type_hint == ArtifactType.NUMERIC
            or output_artifact_type_hint == ArtifactType.BOOL
        )

        operator_id = generate_uuid()
        output_artifact_id = generate_uuid()
        artifact_name = default_artifact_name_from_op_name(op_name)

        create_metric_or_check_artifact(
            dag=self._dag,
            op=Operator(
                id=operator_id,
                name=op_name,
                description=op_description,
                spec=op_spec,
                inputs=[self._artifact_id],
                outputs=[output_artifact_id],
            ),
            output_artifacts=[
                ArtifactMetadata(
                    id=output_artifact_id,
                    name=artifact_name,
                    type=output_artifact_type_hint,
                    explicitly_named=False,
                )
            ],
        )

        if execution_mode == ExecutionMode.EAGER:
            # Issue preview request since this is an eager execution.
            artifact = artifact_utils.preview_artifact(self._dag, output_artifact_id)

            assert isinstance(artifact, numeric_artifact.NumericArtifact) or isinstance(
                artifact, bool_artifact.BoolArtifact
            )
            return artifact
        else:
            # We are in lazy mode.
            if output_artifact_type_hint == ArtifactType.NUMERIC:
                return numeric_artifact.NumericArtifact(self._dag, output_artifact_id)
            else:
                return bool_artifact.BoolArtifact(self._dag, output_artifact_id)
