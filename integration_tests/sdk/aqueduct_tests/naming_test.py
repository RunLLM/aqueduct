import pytest
from aqueduct.error import ArtifactNotFoundException, InvalidUserActionException

from ..shared.data_objects import DataObject
from ..shared.flow_helpers import publish_flow_test
from .extract import extract
from .test_functions.simple.model import (
    dummy_model,
    dummy_model_2,
    dummy_sentiment_model,
    dummy_sentiment_model_multiple_input,
)


def test_extract_with_default_name_collision(client, data_integration):
    # In the case where no explicit name is supplied, we expect new extract
    # operators to always be created.
    table_artifact_1 = extract(data_integration, DataObject.SENTIMENT)
    table_artifact_2 = extract(data_integration, DataObject.SENTIMENT)

    assert table_artifact_1.name() == "%s query artifact" % data_integration.name()
    assert table_artifact_2.name() == "%s query (1) artifact" % data_integration.name()

    fn_artifact = dummy_sentiment_model_multiple_input(table_artifact_1, table_artifact_2)
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
    table_artifact_1 = extract(data_integration, DataObject.SENTIMENT, op_name="sql query")

    fn_artifact = dummy_sentiment_model(table_artifact_1)

    table_artifact_2 = extract(data_integration, DataObject.SENTIMENT, op_name="sql query")
    assert table_artifact_2.name() == "sql query artifact"

    # Cannot preview an artifact with a dependency that has been deleted,
    # since it itself would have been removed from the dag.
    with pytest.raises(ArtifactNotFoundException):
        fn_artifact.get()

    # Cannot run a function on an artifact that has already been overwritten.
    with pytest.raises(ArtifactNotFoundException):
        _ = dummy_sentiment_model(table_artifact_1)

    fn_artifact = dummy_sentiment_model(table_artifact_2)
    fn_df = fn_artifact.get()
    assert list(fn_df) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
        "positivity",
    ]
    assert fn_df.shape[0] == 100


def test_extract_with_custom_artifact(client, data_integration, engine, flow_name):
    output = extract(data_integration, DataObject.SENTIMENT, output_name="hotel reviews")
    assert output.name() == "hotel reviews"

    flow = publish_flow_test(client, artifacts=output, engine=engine, name=flow_name())
    assert flow.latest().artifact("hotel reviews").get().equals(output.get())

    # Cannot name an output artifact the same as an existing one.
    with pytest.raises(InvalidUserActionException, match="has already been created locally"):
        extract(data_integration, DataObject.SENTIMENT, output_name="hotel reviews")


def test_extract_with_default_artifact_name_collision(client, data_integration):
    # The default output artifact name can also collide with an existing artifact!
    client.create_param("sql query artifact", default=123)
    with pytest.raises(
        InvalidUserActionException,
        match="has already been created locally. Artifact names must be unique.",
    ):
        extract(data_integration, DataObject.SENTIMENT, op_name="sql query")


def test_function_with_name_collision(client, data_integration):
    """Colliding functions are always overwritten."""
    table_artifact = extract(data_integration, DataObject.SENTIMENT, op_name="sql query")

    # There's not an easy way to programmatically change the function, so lets
    # just run the same function twice and check that the latest one wins.
    dummy_fn_artifact_old = dummy_model(table_artifact)
    dummy_fn_artifact_new = dummy_model(table_artifact)

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
    table_artifact = extract(data_integration, DataObject.SENTIMENT, op_name="sql query")

    # SQL collides with existing function
    _ = dummy_sentiment_model(table_artifact)
    with pytest.raises(InvalidUserActionException):
        _ = extract(data_integration, DataObject.SENTIMENT, op_name="dummy_sentiment_model")
