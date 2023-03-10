from __future__ import annotations

from typing import Any

from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.transform import to_artifact_class
from aqueduct.constants.enums import OperatorType
from aqueduct.error import InvalidUserActionException, InvalidUserArgumentException
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.operators import Operator, OperatorSpec, get_operator_type
from aqueduct.utils.dag_deltas import AddOperatorDelta, apply_deltas_to_dag, ReplaceOperatorDelta
from aqueduct.utils.type_inference import infer_artifact_type
from aqueduct.utils.utils import construct_param_spec, generate_uuid


def operator_is_implicitly_created_param(op: Operator) -> bool:
    if get_operator_type(op) != OperatorType.PARAM:
        return False
    assert op.spec.param is not None
    return op.spec.param.implicitly_created


def check_implicit_param_name(dag: DAG, candidate_name: str, op_name: str) -> bool:
    """We will either error or overwrite the colliding parameter, if it is consumed by the same op_name.

    Returns whether this is a new parameter or we're overwriting an existing one.
    """
    colliding_artifact = dag.get_artifact_by_name(candidate_name)
    colliding_param_op = dag.get_param_op_by_name(candidate_name)

    # No collisions.
    if colliding_param_op is None and colliding_artifact is None:
        return False

    # If colliding with both another operator and artifact, check whether we can overwrite.
    if colliding_param_op is not None and operator_is_implicitly_created_param(colliding_param_op):
            assert len(colliding_param_op.outputs) == 1, "Parameter operator must have a single output."
            ops = dag.list_operators(on_artifact_id=colliding_param_op.outputs[0])
            assert len(ops) == 1, "Implicit parameters can only be consumed by a single operator."

            # We only overwrite if it's an exact replacement!
            if op_name == ops[0].name:
                return True

    # Anything else is not salvagable.
    raise InvalidUserActionException(
        """Unable to create parameter `%s`, since there is an existing operator or artifact with the same name."""
        % candidate_name
    )


def create_param_artifact(
    dag: DAG,
    param_name: str,
    default: Any,
    overwrite: bool,
    description: str = "",
    is_implicit: bool = False,
) -> BaseArtifact:
    """Creates a parameter operator and return an artifact that can be fed into other operators.

    Args:
        dag:
            The dag to check for collisions against.
        param_name:
            The name for the parameter.
        overwrite:
            Whether to overwrite an existing parameter with the same name. If set, we assume such a parameter exists.
        default:
            The default value for the new parameter.
        description:
            A description for the parameter.
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

    if overwrite:
        delta = ReplaceOperatorDelta(op=op, output_artifacts=output_artifacts)
    else:
        delta = AddOperatorDelta(op=op, output_artifacts=output_artifacts)

    apply_deltas_to_dag(
        dag,
        deltas=[delta],
    )
    return to_artifact_class(dag, output_artifact_id, artifact_type, default)
