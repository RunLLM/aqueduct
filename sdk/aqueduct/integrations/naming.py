from typing import Optional, Tuple

from aqueduct.models.dag import DAG
from aqueduct.utils.naming import resolve_op_and_artifact_names


def _resolve_op_and_artifact_name_for_extract(
    dag: DAG,
    op_name: Optional[str],
    default_op_name: str,
    artifact_name: Optional[str],
) -> Tuple[str, str]:
    """For extract operators, if an explicit name is provided, we will overwrite the existing one.

    Otherwise, we'll deduplicate the default operator names.
    """
    candidate_op_name = op_name or default_op_name
    overwrite_existing_op_name = op_name is not None
    op_name, artifact_names = resolve_op_and_artifact_names(
        dag,
        candidate_op_name,
        overwrite_existing_op_name=overwrite_existing_op_name,
        candidate_artifact_names=artifact_name,
    )
    assert len(artifact_names) == 1
    return op_name, artifact_names[0]
