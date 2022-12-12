from __future__ import annotations

from typing import Any

from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.transform import to_artifact_class
from aqueduct.error import InvalidUserArgumentException
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.operators import Operator, OperatorSpec
from aqueduct.utils.dag_deltas import AddOrReplaceOperatorDelta, apply_deltas_to_dag
from aqueduct.utils.type_inference import infer_artifact_type
from aqueduct.utils.utils import construct_param_spec, generate_uuid


def create_param_artifact(
    dag: DAG,
    name: str,
    default: Any,
    description: str = "",
) -> BaseArtifact:
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
