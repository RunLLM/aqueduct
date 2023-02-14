from typing import List, Optional, Tuple, Union

from aqueduct.models.dag import DAG
from aqueduct.utils.dag_deltas import RemoveOperatorDelta, apply_deltas_to_dag


def _artifact_name_from_op_name(op_name: str) -> str:
    return op_name + " artifact"


def _construct_default_output_artifact_names_from_op(op_name: str, num_outputs: int) -> List[str]:
    """The default artifact naming policy is "<op_name> artifact (<optional counter>)".

    In the multi-output case, we deduplicate each, starting the counter at 1.
    """
    artifact_names = [_artifact_name_from_op_name(op_name)]
    if num_outputs == 1:
        return artifact_names

    artifact_names += [
        _artifact_name_from_op_name(op_name) + " (%d)" % i for i in range(1, num_outputs)
    ]
    return artifact_names


def resolve_op_and_artifact_names(
    dag: DAG,
    candidate_op_name: str,
    overwrite_existing_op_name: bool,
    candidate_artifact_names: Optional[Union[str, List[str]]] = None,
    num_outputs: int = 1,
    only_resolve_op_name: bool = False,
) -> Tuple[str, List[str]]:
    """Enforces our unique naming policy on operators and also their output artifacts.

    Operator collisions are resolved. Artifact collisions are not tolerated.

    Args:
        dag:
            The current dag.
        candidate_op_name:
            The operator name that we want to check and potentially deduplicate.
        overwrite_existing_op_name:
            If set, we will overwrite any existing operator with the same name.
            If not set, we will bump the operator name until an available one is found.
        candidate_artifact_names:
            The custom output artifact names the caller wants to assign to the operator
            outputs. If not set, we will use our default artifact naming scheme. Regardless,
            if there already exist artifact's with the same name, we will error.
        num_outputs:
            Must be consistent with `candidate_artifact_names`, if set. The number of
            output artifacts to expected.
        only_resolve_op_name:
            NOTE: only use this if the operator does not produce any artifacts! (eg. saves).
            If set, this will skip the artifact naming validation step. An empty list will be
            returned for the artifact names.
    Returns:
        A tuple of the operator name and artifact name(s), which the caller can use
        to safely attach elements to the DAG.
    """
    # First, resolve the candidate op name. If we can overwrite an existing op name, then
    # don't even bother finding an unallocated op name.
    if not overwrite_existing_op_name:
        prefix = candidate_op_name
        curr_suffix = 1
        while dag.get_operator(with_name=candidate_op_name) is not None:
            candidate_op_name = prefix + " (%d)" % curr_suffix
            curr_suffix += 1

    # If we're overwriting an existing operator, temporarily delete the operator from the dag
    # so that downstream names won't collide with our future naming checks.
    else:
        colliding_op = dag.get_operator(with_name=candidate_op_name)
        if colliding_op is not None:
            dag = apply_deltas_to_dag(
                dag,
                deltas=[
                    RemoveOperatorDelta(op_id=colliding_op.id),
                ],
                make_copy=True,
            )
    op_name = candidate_op_name
    if only_resolve_op_name:
        return op_name, []

    # Second, validate the artifact name(s) does not collide with other artifacts.
    if candidate_artifact_names is None:
        candidate_artifact_names = _construct_default_output_artifact_names_from_op(
            op_name, num_outputs
        )
    assert candidate_artifact_names is not None

    if isinstance(candidate_artifact_names, str):
        candidate_artifact_names = [candidate_artifact_names]
    assert isinstance(candidate_artifact_names, list)
    assert num_outputs == len(candidate_artifact_names)

    for candidate_artifact_name in candidate_artifact_names:
        dag.validate_artifact_name(candidate_artifact_name)
    artifact_names = candidate_artifact_names

    return op_name, artifact_names
