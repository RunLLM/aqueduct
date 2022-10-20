import time

import pandas as pd
import pytest
from aqueduct.error import AqueductError, InvalidDependencyFilePath, InvalidFunctionException
from constants import SENTIMENT_SQL_QUERY
from test_functions.simple.file_dependency_model import (
    model_with_file_dependency,
    model_with_improper_dependency_path,
    model_with_invalid_dependencies,
    model_with_missing_file_dependencies,
    model_with_out_of_package_file_dependency,
)
from test_functions.simple.model import dummy_model, dummy_sentiment_model, dummy_sentiment_model_multiple_input
from utils import get_integration_name

from aqueduct import op


def test_basic_get(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
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


def test_multiple_input_get(client):
    db = client.integration(name=get_integration_name())
    sql_artifact1 = db.sql(name="Query 1", query=SENTIMENT_SQL_QUERY)
    sql_artifact2 = db.sql(name="Query 2", query=SENTIMENT_SQL_QUERY)

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


def test_basic_file_dependencies(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)

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


def test_invalid_file_dependencies(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)

    with pytest.raises(AqueductError):
        model_with_invalid_dependencies(sql_artifact)

    with pytest.raises(AqueductError):
        model_with_missing_file_dependencies(sql_artifact)

    with pytest.raises(InvalidFunctionException):
        model_with_improper_dependency_path(sql_artifact)

    with pytest.raises(InvalidDependencyFilePath):
        model_with_out_of_package_file_dependency(sql_artifact)


def test_preview_artifact_caching(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)

    @op
    def slow_fn(df):
        time.sleep(10)
        return df

    @op
    def noop(df):
        return df

    # Check that the first run will take a while, but the second run will happen much faster.
    start = time.time()
    slow_output = slow_fn(sql_artifact)
    duration = time.time() - start
    assert duration > 10

    start = time.time()
    _ = noop(slow_output)
    assert time.time() - start < duration


def test_table_with_non_string_column_name(client):
    @op
    def bad_return():
        return pd.DataFrame([0, 1, 2, 3], columns=[123])

    with pytest.raises(AqueductError):
        bad_return()
