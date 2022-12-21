import pandas as pd
import pytest
from aqueduct.error import AqueductError
from constants import SENTIMENT_SQL_QUERY
from test_metrics.constant.model import constant_metric
from utils import publish_flow_test

from aqueduct import metric


def test_basic_metric(client, data_integration):
    sql_artifact = data_integration.sql(query=SENTIMENT_SQL_QUERY)

    metric = constant_metric(sql_artifact)
    assert metric.get() == 17.5


def test_metric_bound(client, data_integration):
    sql_artifact = data_integration.sql(query=SENTIMENT_SQL_QUERY)

    metric = constant_metric(sql_artifact)
    check_artifact = metric.bound(upper=100)
    assert check_artifact.get()

    check_artifact = metric.bound(lower=100)
    assert not check_artifact.get()

    with pytest.raises(AqueductError):
        _ = metric.bound(lower="100")

    with pytest.raises(AqueductError):
        _ = metric.bound(lower=100, upper=200)

    check_artifact = metric.bound(equal=17.5)
    assert check_artifact.get()

    check_artifact = metric.bound(notequal=17.5)
    assert not check_artifact.get()


def test_register_metric(client, flow_name, data_integration, engine):
    sql_artifact = data_integration.sql(query=SENTIMENT_SQL_QUERY)
    metric_artifact = constant_metric(sql_artifact)
    publish_flow_test(
        client,
        name=flow_name(),
        artifacts=[sql_artifact, metric_artifact],
        engine=engine,
    )


@metric()
def metric_with_multiple_inputs(df1, m, df2):
    if not isinstance(df1, pd.DataFrame) or not isinstance(df2, pd.DataFrame):
        raise Exception(
            "Expected dataframes as first and third args, got %s and %s"
            % (type(df1).__name__, type(df2).__name__)
        )
    if not isinstance(m, float):
        raise Exception("Expected float as input to check, got %s" % type(m).__name__)
    return m + 10


def test_metric_mixed_inputs(client, flow_name, data_integration, engine):
    sql1 = data_integration.sql(query=SENTIMENT_SQL_QUERY)
    sql2 = data_integration.sql(query=SENTIMENT_SQL_QUERY)
    metric_input = constant_metric(sql1)

    metric_output = metric_with_multiple_inputs(sql1, metric_input, sql2)
    assert metric_output.get() == 27.5

    publish_flow_test(
        client,
        metric_output,
        name=flow_name(),
        engine=engine,
    )
