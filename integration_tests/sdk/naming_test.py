import pytest
from aqueduct.error import ArtifactNotFoundException, InvalidUserActionException
from constants import SENTIMENT_SQL_QUERY
from test_functions.simple.model import (
    dummy_model,
    dummy_model_2,
    dummy_sentiment_model,
    dummy_sentiment_model_multiple_input,
)


def test_extract_with_default_name_collision(client, data_integration):
    # In the case where no explicit name is supplied, we expect new extract
    # operators to always be created.
    sql_artifact_1 = data_integration.sql(query=SENTIMENT_SQL_QUERY)
    sql_artifact_2 = data_integration.sql(query=SENTIMENT_SQL_QUERY)

    assert sql_artifact_1.name() == "%s query 1 artifact" % data_integration
    assert sql_artifact_2.name() == "%s query 2 artifact" % data_integration

    fn_artifact = dummy_sentiment_model_multiple_input(sql_artifact_1, sql_artifact_2)
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


def test_extract_with_explicit_name_collision(client, data_integration):
    # In the case where an explicit name is supplied, we will overwrite any colliding ops.
    sql_artifact_1 = data_integration.sql(query=SENTIMENT_SQL_QUERY, name="sql query")

    fn_artifact = dummy_sentiment_model(sql_artifact_1)

    sql_artifact_2 = data_integration.sql(query=SENTIMENT_SQL_QUERY, name="sql query")
    assert sql_artifact_2.name() == "sql query artifact"

    # Cannot preview an artifact with a dependency that has been deleted,
    # since it itself would have been removed from the dag.
    with pytest.raises(ArtifactNotFoundException):
        fn_artifact.get()

    # Cannot run a function on an artifact that has already been overwritten.
    with pytest.raises(ArtifactNotFoundException):
        _ = dummy_sentiment_model(sql_artifact_1)

    fn_artifact = dummy_sentiment_model(sql_artifact_2)
    fn_df = fn_artifact.get()
    assert list(fn_df) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
        "positivity",
    ]
    assert fn_df.shape[0] == 100


def test_function_with_name_collision(client, data_integration):
    """Colliding functions are always overwritten."""
    sql_artifact = data_integration.sql(query=SENTIMENT_SQL_QUERY, name="sql query")

    # There's not an easy way to programmatically change the function, so lets
    # just run the same function twice and check that the latest one wins.
    dummy_fn_artifact_old = dummy_model(sql_artifact)
    dummy_fn_artifact_new = dummy_model(sql_artifact)

    with pytest.raises(ArtifactNotFoundException):
        dummy_fn_artifact_old.get()

    fn_df = dummy_fn_artifact_new.get()
    assert list(fn_df) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
        "newcol",
    ]
    assert fn_df.shape[0] == 100


def test_naming_collision_with_different_types(client, data_integration):
    # An overwrite is invalid because the operators are of different types.
    sql_artifact = data_integration.sql(query=SENTIMENT_SQL_QUERY, name="sql query")

    # Function collides with existing sql artifact
    _ = data_integration.sql(query=SENTIMENT_SQL_QUERY, name="dummy_model")
    with pytest.raises(InvalidUserActionException):
        dummy_model(sql_artifact)

    # SQL collides with existing function
    _ = dummy_sentiment_model(sql_artifact)
    with pytest.raises(InvalidUserActionException):
        _ = data_integration.sql(query=SENTIMENT_SQL_QUERY, name="dummy_sentiment_model")


def test_naming_collision_with_dependency(client, data_integration):
    # Overwrite is invalid when the operator being replaced is an upstream dependency.
    sql_artifact = data_integration.sql(query=SENTIMENT_SQL_QUERY, name="sentiment_model")
    dummy_model_output = dummy_model(sql_artifact)
    dummy_model_2_output = dummy_model_2(dummy_model_output)

    with pytest.raises(InvalidUserActionException):
        _ = dummy_model(dummy_model_2_output)
