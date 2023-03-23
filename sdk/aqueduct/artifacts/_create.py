from __future__ import annotations

from typing import List

from aqueduct.constants.enums import OperatorType
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.operators import Operator, get_operator_type
from aqueduct.utils.dag_deltas import (
    AddOperatorDelta,
    DAGDelta,
    RemoveOperatorDelta,
    apply_deltas_to_dag,
)


def create_metric_or_check_artifact(
    dag: DAG,
    op: Operator,
    output_artifacts: List[ArtifactMetadata],
) -> None:
    """Adds a metric/check operator to the DAG.

    Replaces an existing metric/check iff they have the same name, type, and input artifacts.
    """
    assert get_operator_type(op) in [
        OperatorType.METRIC,
        OperatorType.CHECK,
        OperatorType.SYSTEM_METRIC,
    ]

    deltas: List[DAGDelta] = []
    op_to_overwrite = dag.get_colliding_metric_or_check(op)
    if op_to_overwrite is not None:
        deltas.append(
            RemoveOperatorDelta(
                op_id=op_to_overwrite.id,
            ),
        )

    deltas.append(
        AddOperatorDelta(
            op=op,
            output_artifacts=output_artifacts,
        )
    )
    apply_deltas_to_dag(
        dag,
        deltas=deltas,
    )
