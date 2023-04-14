import uuid
from typing import Any, Dict, Union

from aqueduct.artifacts import bool_artifact, numeric_artifact
from aqueduct.artifacts import preview as artifact_utils
from aqueduct.artifacts._create import create_metric_or_check_artifact
from aqueduct.constants.enums import ArtifactType, ExecutionMode
from aqueduct.constants.metrics import SYSTEM_METRICS_INFO
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.operators import Operator, OperatorSpec, SystemMetricSpec
from aqueduct.utils.naming import default_artifact_name_from_op_name
from aqueduct.utils.utils import generate_uuid

from aqueduct import globals


class SystemMetricMixin:
    """A mixin class for the system_metric function. This is used by GenericArtifacts and TableArtifacts."""

    def list_system_metrics(self) -> Dict[str, Any]:
        """Returns a dictionary of all system metrics available on the table artifact.
        These system metrics can be set via the invoking the system_metric() method the table.

        Returns:
            A list of available system metrics on a table
        """
        return SYSTEM_METRICS_INFO

    def _system_metric_helper(
        self, dag: DAG, artifact_id: uuid.UUID, metric_name: str, lazy: bool
    ) -> numeric_artifact.NumericArtifact:
        execution_mode = ExecutionMode.EAGER if not lazy else ExecutionMode.LAZY

        operator = dag.must_get_operator(with_output_artifact_id=artifact_id)
        system_metric_description, system_metric_unit = SYSTEM_METRICS_INFO[metric_name]
        system_metric_name = "%s %s(%s) metric" % (operator.name, metric_name, system_metric_unit)
        op_spec = OperatorSpec(system_metric=SystemMetricSpec(metric_name=metric_name))
        new_artifact = self._apply_operator_to_table(
            dag,
            artifact_id,
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
        dag: DAG,
        artifact_id: uuid.UUID,
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
            dag=dag,
            op=Operator(
                id=operator_id,
                name=op_name,
                description=op_description,
                spec=op_spec,
                inputs=[artifact_id],
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
            artifact = artifact_utils.preview_artifact(dag, output_artifact_id)

            assert isinstance(artifact, numeric_artifact.NumericArtifact) or isinstance(
                artifact, bool_artifact.BoolArtifact
            )
            return artifact
        else:
            # We are in lazy mode.
            if output_artifact_type_hint == ArtifactType.NUMERIC:
                return numeric_artifact.NumericArtifact(dag, output_artifact_id)
            else:
                return bool_artifact.BoolArtifact(dag, output_artifact_id)
