from typing import Any, Dict, List, Mapping, Optional, Union

from aqueduct.constants.enums import OperatorType
from aqueduct.error import AqueductError
from aqueduct.models.dag import DAG
from aqueduct.models.operators import Operator


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


def _get_checks_for_op(op: Operator, dag: DAG) -> List[Operator]:
    check_operators = []
    for artf in op.outputs:
        check_operators.extend(
            dag.list_operators(
                filter_to=[OperatorType.CHECK],
                on_artifact_id=artf,
            )
        )
    return check_operators


def _get_metrics_for_op(op: Operator, dag: DAG) -> List[Operator]:
    metric_operators = []
    for artf in op.outputs:
        metric_operators.extend(
            dag.list_operators(
                filter_to=[OperatorType.METRIC],
                on_artifact_id=artf,
            )
        )
    return metric_operators


def _get_description_for_check(check: Operator) -> Dict[str, str]:
    check_spec = check.spec.check
    if check_spec:
        level = check_spec.level
    else:
        raise AqueductError("Check artifact malformed.")
    return {
        "Label": check.name,
        "Description": check.description,
        "Level": level,
    }


def _get_description_for_metric(
    metric: Operator, dag: DAG
) -> Dict[str, Union[str, List[Mapping[str, Any]]]]:
    metric_spec = metric.spec.metric
    if metric_spec:
        granularity = metric_spec.function.granularity
    else:
        raise AqueductError("Metric artifact malformed.")
    return {
        "Label": metric.name,
        "Description": metric.description,
        "Granularity": granularity,
        "Checks": [
            _get_description_for_check(check_op) for check_op in _get_checks_for_op(metric, dag)
        ],
        "Metrics": [
            _get_description_for_metric(metric_op, dag)
            for metric_op in _get_metrics_for_op(metric, dag)
        ],
    }
