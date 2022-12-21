from typing import Any, Dict, List, Mapping, Union

from aqueduct.error import AqueductError
from aqueduct.models.dag import DAG
from aqueduct.models.operators import Operator


def get_readable_description_for_check(check: Operator) -> Dict[str, str]:
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


def get_readable_description_for_metric(
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
            get_readable_description_for_check(check_op)
            for check_op in dag.list_checks_for_operator(metric)
        ],
        "Metrics": [
            get_readable_description_for_metric(metric_op, dag)
            for metric_op in dag.list_metrics_for_operator(metric)
        ],
    }
