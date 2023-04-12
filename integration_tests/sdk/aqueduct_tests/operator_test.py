import pandas as pd
import pytest
from aqueduct.decorator import to_operator
from aqueduct.error import ArtifactNotFoundException

from aqueduct import op
from sdk.aqueduct_tests.test_function import dummy_sentiment_model_function

from ..shared.data_objects import DataObject
from .extract import extract


def test_to_operator_local_function(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    @op()
    def dummy_sentiment_model(df):
        if isinstance(df, pd.DataFrame):
            df["POSITIVITY"] = 123
        else:
            from pyspark.sql.functions import lit

            df = df.withColumn("POSITIVITY", lit(123.0))

        return df

    def dummy_sentiment_model_func(df):
        if isinstance(df, pd.DataFrame):
            df["POSITIVITY"] = 123
        else:
            from pyspark.sql.functions import lit

            df = df.withColumn("POSITIVITY", lit(123.0))

        return df

    output_artifact_from_decorator = dummy_sentiment_model(table_artifact)
    df_normal = output_artifact_from_decorator.get()
    output_artifact_from_to_operator = to_operator(dummy_sentiment_model_func)(table_artifact)
    df_func = output_artifact_from_to_operator.get()

    assert df_normal["POSITIVITY"].equals(df_func["POSITIVITY"])


def test_operator_reuse_chain(data_integration):
    """Tests reusing the same operator when it is chained together by a dependency."""
    wine_artifact = extract(data_integration, DataObject.WINE)

    @op
    def noop_1(df):
        return df

    @op
    def noop_2(df):
        return df

    a = noop_1(wine_artifact)
    b = noop_2(a)
    c = noop_1(b)

    _ = a.get()
    _ = b.get()
    _ = c.get()

    assert a.name() == "noop_1 artifact"
    assert b.name() == "noop_2 artifact"
    assert c.name() == "noop_1 artifact"

    # All the artifacts still work.
    assert a.get().equals(wine_artifact.get())
    assert b.get().equals(wine_artifact.get())
    assert c.get().equals(wine_artifact.get())


# TODO(ENG-1470): This doesn't work in pytest, but is fine in a jupyter notebook.
# def test_to_operator_imported_function(client, data_integration):
#     table_artifact = extract(data_integration, DataObject.SENTIMENT)
#
#     @op(file_dependencies=["test_function.py"])
#     def decorated_func(df):
#         df = dummy_sentiment_model_function(df)
#         return df
#
#     df_decorate = decorated_func(table_artifact).get()
#     df_function = to_operator(
#         dummy_sentiment_model_function, file_dependencies=["test_function.py"]
#     )(table_artifact).get()
#
#     assert df_decorate["positivity"].equals(df_function["positivity"])
