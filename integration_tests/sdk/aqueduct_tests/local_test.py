import pytest
from pandas import DataFrame
from pandas._testing import assert_frame_equal

from ..shared.data_objects import DataObject
from .checks_test import success_on_single_table_input
from .extract import extract
from .test_functions.simple.model import dummy_sentiment_model, dummy_sentiment_model_multiple_input
from .test_metrics.constant.model import constant_metric


def test_local_operator(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)
    output_cloud = output_artifact.get()

    output_local = dummy_sentiment_model.local(table_artifact)
    assert output_cloud.count()[0] == output_local.count()[0]
    output_cloud.columns = output_cloud.columns.str.lower()
    output_local.columns = output_local.columns.str.lower()
    assert_frame_equal(output_cloud, output_local)


def test_local_metric(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    metric = constant_metric(table_artifact)
    assert metric.get() == 17.5
    assert constant_metric.local(table_artifact) == 17.5


def test_local_check(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    check = success_on_single_table_input
    assert check(table_artifact)
    assert check.local(table_artifact)


def test_local_dataframe_input(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    output_cloud = dummy_sentiment_model(table_artifact).get()
    output_local = dummy_sentiment_model.local(table_artifact.get())
    assert type(output_local) is DataFrame
    output_cloud.columns = output_cloud.columns.str.lower()
    output_local.columns = output_local.columns.str.lower()
    assert_frame_equal(output_cloud, output_local)


def test_local_on_multiple_inputs(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    table_artifact2 = extract(data_integration, DataObject.SENTIMENT)

    output_cloud = dummy_sentiment_model_multiple_input(table_artifact, table_artifact2).get()
    output_local = dummy_sentiment_model_multiple_input.local(table_artifact, table_artifact2)
    assert type(output_local) is DataFrame
    output_cloud.columns = output_cloud.columns.str.lower()
    output_local.columns = output_local.columns.str.lower()
    assert_frame_equal(output_cloud, output_local)
