from __future__ import annotations

from typing import Any

from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.transform import to_artifact_class
from aqueduct.error import InvalidUserArgumentException
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.operators import Operator, OperatorSpec, get_operator_type
from aqueduct.utils.dag_deltas import AddOperatorDelta, apply_deltas_to_dag
from aqueduct.utils.type_inference import infer_artifact_type
from aqueduct.utils.utils import construct_param_spec, generate_uuid


def create_param_artifact(
    dag: DAG,
    param_name: str,
    default: Any,
    description: str = "",
    is_implicit: bool = False,
) -> BaseArtifact:
    """Creates a parameter operator and return an artifact that can be fed into other operators.

    Args:
        dag:
            The dag to check for collisions against.
        param_name:
            The name for the parameter.
        default:
            The default value for the new parameter.
        description:
            A description for the parameter.
        is_implicit:
            Whether this parameter is implicitly created.
    """
    if default is None:
        raise InvalidUserArgumentException("Parameter default value cannot be None.")

    artifact_type = infer_artifact_type(default)
    param_spec = construct_param_spec(default, artifact_type, is_implicit=is_implicit)
    operator_id = generate_uuid()
    output_artifact_id = generate_uuid()

    op = Operator(
        id=operator_id,
        name=param_name,
        description=description,
        spec=OperatorSpec(param=param_spec),
        inputs=[],
        outputs=[output_artifact_id],
    )
    output_artifacts = [
        ArtifactMetadata(
            id=output_artifact_id,
            name=param_name,
            type=artifact_type,
        ),
    ]

    apply_deltas_to_dag(
        dag,
        deltas=[
            AddOperatorDelta(
                op=op,
                output_artifacts=output_artifacts,
            ),
        ],
    )
    return to_artifact_class(dag, output_artifact_id, artifact_type, default)
