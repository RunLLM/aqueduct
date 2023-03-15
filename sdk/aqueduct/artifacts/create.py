from __future__ import annotations

import uuid
from typing import Any, Optional

from aqueduct.artifacts import bool_artifact, generic_artifact, numeric_artifact, table_artifact
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.constants.enums import ArtifactType
from aqueduct.error import InvalidUserArgumentException
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.operators import Operator, OperatorSpec
from aqueduct.utils.dag_deltas import AddOperatorDelta, apply_deltas_to_dag
from aqueduct.utils.naming import sanitize_artifact_name
from aqueduct.utils.type_inference import infer_artifact_type
from aqueduct.utils.utils import construct_param_spec, generate_uuid


def to_artifact_class(
    dag: DAG,
    artifact_id: uuid.UUID,
    artifact_type: ArtifactType = ArtifactType.UNTYPED,
    content: Optional[Any] = None,
) -> BaseArtifact:
    if artifact_type == ArtifactType.TABLE:
        return table_artifact.TableArtifact(
            dag,
            artifact_id,
            content,
        )
    elif artifact_type == ArtifactType.NUMERIC:
        return numeric_artifact.NumericArtifact(dag, artifact_id, content)
    elif artifact_type == ArtifactType.BOOL:
        return bool_artifact.BoolArtifact(dag, artifact_id, content)
    else:
        return generic_artifact.GenericArtifact(dag, artifact_id, artifact_type, content)


def create_param_artifact(
    dag: DAG,
    param_name: str,
    default: Any,
    description: str,
    explicitly_named: bool,
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
        explicitly_named:
            Whether this parameter was explicitly created with `client.create_param()`.
    """
    if default is None:
        raise InvalidUserArgumentException("Parameter default value cannot be None.")

    artifact_type = infer_artifact_type(default)
    param_spec = construct_param_spec(default, artifact_type)
    operator_id = generate_uuid()
    output_artifact_id = generate_uuid()

    apply_deltas_to_dag(
        dag,
        deltas=[
            AddOperatorDelta(
                op=Operator(
                    id=operator_id,
                    name=param_name,
                    description=description,
                    spec=OperatorSpec(param=param_spec),
                    inputs=[],
                    outputs=[output_artifact_id],
                ),
                output_artifacts=[
                    ArtifactMetadata(
                        id=output_artifact_id,
                        name=sanitize_artifact_name(param_name),
                        type=artifact_type,
                        explicitly_named=explicitly_named,
                    ),
                ],
            ),
        ],
    )
    return to_artifact_class(dag, output_artifact_id, artifact_type, default)
