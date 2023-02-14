from __future__ import annotations

import warnings
from typing import Any, Optional

from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.transform import to_artifact_class
from aqueduct.constants.enums import OperatorType
from aqueduct.error import InvalidUserActionException, InvalidUserArgumentException
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.operators import Operator, OperatorSpec, get_operator_type
from aqueduct.utils.dag_deltas import AddOrReplaceOperatorDelta, apply_deltas_to_dag
from aqueduct.utils.type_inference import infer_artifact_type
from aqueduct.utils.utils import construct_param_spec, generate_uuid


def _operator_is_implicitly_created_param(op: Operator) -> bool:
    if get_operator_type(op) != OperatorType.PARAM:
        return False
    assert op.spec.param is not None
    return op.spec.param.implicitly_created


def _resolve_implicit_param_name(dag: DAG, candidate_name: str, op_name: str) -> bool:
    """We will either error or overwrite the colliding parameter, if it is consumed by the same op_name.

    Returns whether this is a new parameter or we're overwriting an existing one.
    """
    colliding_artifact = dag.get_artifact_by_name(candidate_name)
    colliding_op = dag.get_operator(with_name=candidate_name)

    # No collisions.
    if colliding_op is None and colliding_artifact is None:
        return False

    # If colliding with both another operator and artifact, check whether we can overwrite.
    # This is because parameter operator-artifact pairs must have the same name.
    elif colliding_op is not None and colliding_artifact is not None:
        if _operator_is_implicitly_created_param(colliding_op):
            assert len(colliding_op.outputs) == 1, "Parameter operator must have a single output."
            ops = dag.list_operators(on_artifact_id=colliding_op.outputs[0])
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
    candidate_name: str,
    default: Any,
    description: str = "",
    op_name_for_implicit_param: Optional[str] = None,
) -> BaseArtifact:
    """Creates a parameter operator and return an artifact that can be fed into other operators.

    For implicitly created parameters, the naming collision policy is as follows: we will error
    if there exists other operators or artifacts with the same name, unless we are overwriting
    another implicit parameter being used by the same operator. An implicit parameter is named
    "<op_name>:<param_name>".

    Args:
        dag:
            The dag to check for collisions against.
        candidate_name:
            The suggested name for the parameter.
        default:
            The default value for the new parameter.
        description:
            A description for the parameter.
        op_name_for_implicit_param:
            Only set for implicit parameters - the name of the operator that will consume
            this parameter as input.
    """
    if default is None:
        raise InvalidUserArgumentException("Parameter default value cannot be None.")

    param_name = candidate_name

    # Check if the parameter is being created implicitly. An implicit parameter will have the operator
    # name prepended to it.
    is_implicit = op_name_for_implicit_param is not None
    if is_implicit:
        assert op_name_for_implicit_param is not None  # for mypy

        param_name = op_name_for_implicit_param + ":" + param_name
        is_overwrite = _resolve_implicit_param_name(
            dag,
            param_name,
            op_name_for_implicit_param,
        )
        if not is_overwrite:
            warnings.warn(
                """Input to function argument `%s` is not an artifact type. We have implicitly created a parameter named `%s` and your input will be used as its default value. This parameter will be used when running the function."""
                % (param_name, param_name)
            )
    else:
        colliding_op = dag.get_operator(with_name=param_name)
        if colliding_op is not None and _operator_is_implicitly_created_param(colliding_op):
            raise InvalidUserActionException(
                """Unable to create parameter `%s`, since there is an implicitly created parameter with the same name. If the old parameter is not longer relevant, you can remove it with `client.delete_param()` and rerun this operation. Otherwise, you'll need to rename one of the two. """
                % param_name,
            )

    artifact_type = infer_artifact_type(default)
    param_spec = construct_param_spec(default, artifact_type, is_implicit=is_implicit)

    operator_id = generate_uuid()
    output_artifact_id = generate_uuid()
    apply_deltas_to_dag(
        dag,
        deltas=[
            AddOrReplaceOperatorDelta(
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
                        name=param_name,
                        type=artifact_type,
                    ),
                ],
            )
        ],
    )

    return to_artifact_class(dag, output_artifact_id, artifact_type, default)
