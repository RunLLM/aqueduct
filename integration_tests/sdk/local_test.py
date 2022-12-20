from checks_test import success_on_single_table_input
from constants import SENTIMENT_SQL_QUERY
from pandas import DataFrame
from pandas._testing import assert_frame_equal
from test_functions.simple.model import dummy_sentiment_model, dummy_sentiment_model_multiple_input
from test_metrics.constant.model import constant_metric


def test_local_operator(client, data_integration):
    sql_artifact = data_integration.sql(query=SENTIMENT_SQL_QUERY)
    output_artifact = dummy_sentiment_model(sql_artifact)
    output_cloud = output_artifact.get()

    output_local = dummy_sentiment_model.local(sql_artifact)
    assert output_cloud.count()[0] == output_local.count()[0]
    assert_frame_equal(output_cloud, output_local)


def test_local_metric(client, data_integration):
    sql_artifact = data_integration.sql(query=SENTIMENT_SQL_QUERY)

    metric = constant_metric(sql_artifact)
    assert metric.get() == 17.5
    assert constant_metric.local(sql_artifact) == 17.5


def test_local_check(client, data_integration):
    sql_artifact = data_integration.sql(query=SENTIMENT_SQL_QUERY)
    check = success_on_single_table_input
    assert check(sql_artifact)
    assert check.local(sql_artifact)


def test_local_dataframe_input(client, data_integration):
    sql_artifact = data_integration.sql(query=SENTIMENT_SQL_QUERY)
    output_cloud = dummy_sentiment_model(sql_artifact).get()
    output_local = dummy_sentiment_model.local(sql_artifact.get())
    assert type(output_local) is DataFrame
    assert_frame_equal(output_cloud, output_local)


def test_local_on_multiple_inputs(client, data_integration):
    sql_artifact = data_integration.sql(query=SENTIMENT_SQL_QUERY)
    sql_artifact2 = data_integration.sql(query=SENTIMENT_SQL_QUERY)
    output_cloud = dummy_sentiment_model_multiple_input(sql_artifact, sql_artifact2).get()

    output_local = dummy_sentiment_model_multiple_input.local(sql_artifact, sql_artifact2)
    assert type(output_local) is DataFrame
    assert_frame_equal(output_cloud, output_local)
