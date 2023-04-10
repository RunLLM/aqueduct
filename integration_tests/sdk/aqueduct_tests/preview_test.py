import pandas as pd
import pytest
from aqueduct.error import AqueductError, InvalidDependencyFilePath, InvalidFunctionException

from aqueduct import global_config, op

from ..shared.data_objects import DataObject
from .extract import extract
from .test_functions.simple.file_dependency_model import (
    model_with_file_dependency,
    model_with_improper_dependency_path,
    model_with_invalid_dependencies,
    model_with_missing_file_dependencies,
    model_with_out_of_package_file_dependency,
)
from .test_functions.simple.model import (
    dummy_model,
    dummy_sentiment_model,
    dummy_sentiment_model_multiple_input,
)


def test_basic_get(client, data_integration, engine):
    if engine is not None:
        global_config({"engine": engine})

    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    sql_df = table_artifact.get()
    assert list(sql_df) == ["hotel_name", "review_date", "reviewer_nationality", "review"]
    assert sql_df.shape[0] == 100

    output_artifact = dummy_sentiment_model(table_artifact)
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
    table_artifact1 = extract(data_integration, DataObject.SENTIMENT, op_name="Query 1")
    table_artifact2 = extract(data_integration, DataObject.SENTIMENT, op_name="Query 2")

    fn_artifact = dummy_sentiment_model_multiple_input(table_artifact1, table_artifact2)
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
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    output_artifact = model_with_file_dependency(table_artifact)
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
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    with pytest.raises(AqueductError):
        output_artifact = model_with_invalid_dependencies(table_artifact)
        output_artifact.get()

    with pytest.raises(AqueductError):
        output_artifact = model_with_missing_file_dependencies(table_artifact)
        output_artifact.get()

    with pytest.raises(InvalidFunctionException):
        output_artifact = model_with_improper_dependency_path(table_artifact)
        output_artifact.get()

    with pytest.raises(InvalidDependencyFilePath):
        output_artifact = model_with_out_of_package_file_dependency(table_artifact)
        output_artifact.get()


def test_table_with_non_string_column_name(client):
    @op
    def bad_return():
        return pd.DataFrame([0, 1, 2, 3], columns=[123])

    with pytest.raises(AqueductError):
        output_artifact = bad_return()
        output_artifact.get()
