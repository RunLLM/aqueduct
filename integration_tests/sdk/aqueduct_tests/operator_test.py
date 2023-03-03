import pytest
from aqueduct.decorator import to_operator
from aqueduct.error import ArtifactNotFoundException

from aqueduct import op
from sdk.aqueduct_tests.test_function import dummy_sentiment_model_function

from ..shared.data_objects import DataObject
from .extract import extract


def test_to_operator_local_function(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    @op
    def dummy_sentiment_model(df):
        df["positivity"] = 123
        return df

    def dummy_sentiment_model_func(df):
        df["positivity"] = 123
        return df

    output_artifact_from_decorator = dummy_sentiment_model(table_artifact)
    df_normal = output_artifact_from_decorator.get()
    output_artifact_from_to_operator = to_operator(dummy_sentiment_model_func)(table_artifact)
    df_func = output_artifact_from_to_operator.get()

    assert df_normal["positivity"].equals(df_func["positivity"])


def test_operator_reuse(data_integration):
    """Tests reusing the same operator multiple times in a workflow with different
    input artifacts.
    """
    sentiment_artifact = extract(data_integration, DataObject.SENTIMENT)
    wine_artifact = extract(data_integration, DataObject.WINE)

    @op
    def noop(df):
        return df

    noop_artifact_1 = noop(sentiment_artifact)
    noop_artifact_2 = noop(wine_artifact)

    assert noop_artifact_1.name() == "noop artifact"
    assert noop_artifact_2.name() == "noop (1) artifact"

    _ = noop_artifact_1.get()
    _ = noop_artifact_2.get()

    @op
    def noop_multiple(df1, df2):
        return df1

    # Tests to check that 2 operators are created because the order of
    # the input artifacts are different
    _ = noop_multiple(sentiment_artifact, wine_artifact).get()
    _ = noop_multiple(wine_artifact, sentiment_artifact).get()


def test_operator_overwrite(data_integration):
    """Tests the cases when an operator should be overwritten instead of being
    reused. This happens if an operator with the same name and input artifacts
    is created.
    """

    @op
    def no_args():
        return 456

    no_args_old = no_args()
    no_args_new = no_args()

    # The operator should be overwritten, since the input artifacts (none)
    # are the same.
    with pytest.raises(ArtifactNotFoundException):
        no_args_old.get()

    _ = no_args_new.get()

    sentiment_artifact = extract(data_integration, DataObject.SENTIMENT)
    wine_artifact = extract(data_integration, DataObject.WINE)

    @op
    def single_args(df):
        return df

    single_args_old = single_args(sentiment_artifact)
    single_args_new = single_args(sentiment_artifact)

    # The operator should be overwritten, since the input artifact
    # is the same.
    with pytest.raises(ArtifactNotFoundException):
        single_args_old.get()

    _ = single_args_new.get()

    assert single_args_new.name() == "single_args artifact"

    @op
    def double_args(df1, df2):
        return df1

    double_args_old = double_args(sentiment_artifact, wine_artifact)
    double_args_new = double_args(sentiment_artifact, wine_artifact)

    # The operator should be overwritten, since the input artifacts
    # are the same.
    with pytest.raises(ArtifactNotFoundException):
        double_args_old.get()

    _ = double_args_new.get()

    assert double_args_new.name() == "double_args artifact"


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
    assert c.name() == "noop_1 (1) artifact"


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
