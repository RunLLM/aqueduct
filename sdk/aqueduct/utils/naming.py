from typing import List

from aqueduct.models.dag import DAG


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
