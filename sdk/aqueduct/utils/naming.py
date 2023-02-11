from typing import List, Optional, Tuple, Union

from aqueduct.models.dag import DAG
from aqueduct.utils.dag_deltas import RemoveOperatorDelta, apply_deltas_to_dag


def _artifact_name_from_op_name(op_name: str) -> str:
    return op_name + " artifact"


def _construct_default_output_artifact_names_from_op(op_name: str, num_outputs: int) -> List[str]:
    """The default artifact naming policy is "<op_name> artifact (<optional counter>)".

    In the multi-output case, we deduplicate by starting the counter at 1.
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
    """TODO"""
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
        candidate_artifact_names = _construct_default_output_artifact_names_from_op(op_name, num_outputs)
    assert candidate_artifact_names is not None

    if isinstance(candidate_artifact_names, str):
        candidate_artifact_names = [candidate_artifact_names]
    assert isinstance(candidate_artifact_names, list)
    assert num_outputs == len(candidate_artifact_names)

    for candidate_artifact_name in candidate_artifact_names:
        dag.validate_artifact_name(candidate_artifact_name)
    artifact_names = candidate_artifact_names

    return op_name, artifact_names
