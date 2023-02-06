from typing import Optional

from aqueduct.error import InvalidUserActionException
from aqueduct.models.dag import DAG


def _generate_extract_op_name(
    dag: DAG,
    integration_name: str,
    name: Optional[str],
) -> str:
    """
    Generates name for extract operators to avoid operators with the same name.

    Arguments:
        dag:
            DAG that operator will be a part of.
        integration_name:
            Name of integration to run extract on.
        name:
            Optinally provided operator name.
    Returns:
        Name for extract operator.
    """

    op_name = name

    default_op_prefix = "%s query" % integration_name
    default_op_index = 1
    while op_name is None:
        candidate_op_name = default_op_prefix + " %d" % default_op_index
        colliding_op = dag.get_operator(with_name=candidate_op_name)
        if colliding_op is None:
            op_name = candidate_op_name  # break out of the loop!
        default_op_index += 1

    assert op_name is not None

    return op_name


def _validate_artifact_name(dag: DAG, op_name: str, artifact_name: str) -> None:
    """Checks that the proposed artifact name is unique, expect in the case where
    we are overwriting the colliding operator - artifact pair.
    """
    existing_op = dag.get_operator(with_name=op_name)
    existing_artifact = dag.get_artifact_by_name(artifact_name)

    if existing_artifact is not None:
        # If we are overwriting an existing operator, further check that this overwrite
        # will detach the colliding artifact, thus preserving dag uniqueness.
        if existing_op is not None and len(existing_op.outputs) == 1:
            # TODO(ENG-2399): This is overly restrictive. We should be checking if the colliding
            #  artifact is downstream of the operator being overwritten.
            if existing_artifact == dag.get_artifact(existing_op.outputs[0]):
                return

        raise InvalidUserActionException(
            "Artifact with name `%s` has already been created locally. Artifact names must be unique."
            % artifact_name,
        )
