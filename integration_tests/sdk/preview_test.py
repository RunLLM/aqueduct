import pandas as pd
import pytest
from aqueduct.constants.enums import RuntimeType, ServiceType
from aqueduct.error import AqueductError, InvalidDependencyFilePath, InvalidFunctionException
from constants import SENTIMENT_SQL_QUERY
from test_functions.simple.file_dependency_model import (
    model_with_file_dependency,
    model_with_improper_dependency_path,
    model_with_invalid_dependencies,
    model_with_missing_file_dependencies,
    model_with_out_of_package_file_dependency,
)
from test_functions.simple.model import (
    dummy_model,
    dummy_sentiment_model,
    dummy_sentiment_model_multiple_input,
)

from aqueduct import global_config, op


def test_basic_get(client, data_integration):
    sql_artifact = data_integration.sql(query=SENTIMENT_SQL_QUERY)
    sql_df = sql_artifact.get()
    assert list(sql_df) == ["hotel_name", "review_date", "reviewer_nationality", "review"]
    assert sql_df.shape[0] == 100

    output_artifact = dummy_sentiment_model(sql_artifact)
    output_df = output_artifact.get()
    assert list(output_df) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
        "positivity",
    ]
    assert output_df.shape[0] == 100


def test_multiple_input_get(client, data_integration):
    sql_artifact1 = data_integration.sql(name="Query 1", query=SENTIMENT_SQL_QUERY)
    sql_artifact2 = data_integration.sql(name="Query 2", query=SENTIMENT_SQL_QUERY)

    fn_artifact = dummy_sentiment_model_multiple_input(sql_artifact1, sql_artifact2)
    fn_df = fn_artifact.get()

    assert list(fn_df) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
        "positivity",
        "positivity_2",
    ]
    assert fn_df.shape[0] == 100

    output_artifact = dummy_model(fn_artifact)
    output_df = output_artifact.get()
    assert list(output_df) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
        "positivity",
        "positivity_2",
        "newcol",
    ]
    assert fn_df.shape[0] == 100


def test_basic_file_dependencies(client, data_integration):
    sql_artifact = data_integration.sql(query=SENTIMENT_SQL_QUERY)

    output_artifact = model_with_file_dependency(sql_artifact)
    output_df = output_artifact.get()
    assert list(output_df) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
        "newcol",
    ]
    assert output_df.shape[0] == 100


def test_invalid_file_dependencies(client, data_integration):
    sql_artifact = data_integration.sql(query=SENTIMENT_SQL_QUERY)

    with pytest.raises(AqueductError):
        model_with_invalid_dependencies(sql_artifact)

    with pytest.raises(AqueductError):
        model_with_missing_file_dependencies(sql_artifact)

    with pytest.raises(InvalidFunctionException):
        model_with_improper_dependency_path(sql_artifact)

    with pytest.raises(InvalidDependencyFilePath):
        model_with_out_of_package_file_dependency(sql_artifact)


def test_table_with_non_string_column_name(client):
    @op
    def bad_return():
        return pd.DataFrame([0, 1, 2, 3], columns=[123])

    with pytest.raises(AqueductError):
        bad_return()


@pytest.mark.enable_only_for_engine_type(ServiceType.K8S, ServiceType.LAMBDA)
def test_basic_get_by_engine(client, data_integration, engine):
    global_config({"engine": engine})
    sql_artifact = data_integration.sql(query=SENTIMENT_SQL_QUERY)
    sql_df = sql_artifact.get()
    assert list(sql_df) == ["hotel_name", "review_date", "reviewer_nationality", "review"]
    assert sql_df.shape[0] == 100

    output_artifact = dummy_sentiment_model(sql_artifact)
    integration_info_by_name = client.list_integrations()
    if integration_info_by_name[engine].service == ServiceType.K8S:
        assert output_artifact._dag.engine_config.type == RuntimeType.K8S
    elif integration_info_by_name[engine].service == ServiceType.LAMBDA:
        assert output_artifact._dag.engine_config.type == RuntimeType.K8S
    output_df = output_artifact.get()
    assert list(output_df) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
        "positivity",
    ]
    assert output_df.shape[0] == 100
