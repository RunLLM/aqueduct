from __future__ import annotations

from typing import TYPE_CHECKING, Any, Union

from aqueduct.artifacts.metadata import ArtifactMetadata
from aqueduct.artifacts.utils import to_artifact_class
from aqueduct.dag import DAG
from aqueduct.dag_deltas import AddOrReplaceOperatorDelta, apply_deltas_to_dag
from aqueduct.error import InvalidUserArgumentException
from aqueduct.operators import Operator, OperatorSpec, ParamSpec
from aqueduct.utils import construct_param_spec, generate_uuid, infer_artifact_type

if TYPE_CHECKING:
    from aqueduct.artifacts.bool_artifact import BoolArtifact
    from aqueduct.artifacts.generic_artifact import GenericArtifact
    from aqueduct.artifacts.numeric_artifact import NumericArtifact
    from aqueduct.artifacts.table_artifact import TableArtifact


def create_param(
    dag: DAG,
    name: str,
    default: Any,
    description: str = "",
) -> Union[TableArtifact, NumericArtifact, BoolArtifact, GenericArtifact]:
    """Creates a parameter operator and return an artifact that can be fed into other operators."""
    if default is None:
        raise InvalidUserArgumentException("Parameter default value cannot be None.")

    artifact_type = infer_artifact_type(default)
    param_spec = construct_param_spec(default, artifact_type)

    operator_id = generate_uuid()
    output_artifact_id = generate_uuid()
    apply_deltas_to_dag(
        dag,
        deltas=[
            AddOrReplaceOperatorDelta(
                op=Operator(
                    id=operator_id,
                    name=name,
                    description=description,
                    spec=OperatorSpec(param=param_spec),
                    inputs=[],
                    outputs=[output_artifact_id],
                ),
                output_artifacts=[
                    ArtifactMetadata(
                        id=output_artifact_id,
                        name=name,
                        type=artifact_type,
                    ),
                ],
            )
        ],
    )

    return to_artifact_class(dag, output_artifact_id, artifact_type, default)
