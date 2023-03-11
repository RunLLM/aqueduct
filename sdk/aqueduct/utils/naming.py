from aqueduct.constants.enums import OperatorType
from aqueduct.error import InvalidUserActionException
from aqueduct.models.dag import DAG
from aqueduct.models.operators import Operator, get_operator_type


def default_artifact_name_from_op_name(op_name: str) -> str:
    return op_name + " artifact"


def resolve_artifact_name(dag: DAG, name: str) -> str:
    """TODO: DOCUMENTATION"""
    candidate_name = name
    suffix = 1
    while True:
        colliding_artifact = dag.get_artifact_by_name(candidate_name)
        if colliding_artifact is None:
            break
        candidate_name = name + " (%d)" % suffix
        suffix += 1

    return candidate_name


def operator_is_implicitly_created_param(op: Operator) -> bool:
    if get_operator_type(op) != OperatorType.PARAM:
        return False
    assert op.spec.param is not None
    return op.spec.param.implicitly_created


def resolve_param_name(dag: DAG, name: str, is_implicit: bool) -> str:
    """
    Collision policy:
    - When an implicitly created parameter collides with any artifact, we bump the name with (idx) suffix.
    - When an explicitly created parameter collides:
        - with an existing explicit parameter, we replace it.
        - with an existing implicit parameter, we error.
        - with an existing artifact, we error.
    """
    if is_implicit:
        return resolve_artifact_name(dag, name)
    else:
        # Error if the colliding parameter is implicitly created.
        colliding_param = dag.get_param_op_by_name(name)
        if colliding_param is not None and operator_is_implicitly_created_param(colliding_param):
            raise InvalidUserActionException(
                """Unable to create parameter `%s`, since there is an existing implicit parameter with the same name."""
                % name
            )

        # Error if there are any other artifact name collisions.
        colliding_artifact = dag.get_artifact_by_name(name)
        if colliding_artifact is not None:
            raise InvalidUserActionException(
                """Unable to create parameter `%s`, since there is an existing artifact with the same name."""
                % name
            )

        return name
