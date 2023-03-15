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


def test_extract_with_default_name_collision(client, flow_name, engine, data_integration):
    # In the case where no explicit name is supplied, we expect new extract
    # operators to always be created.
    table_artifact_1 = extract(data_integration, DataObject.SENTIMENT)
    table_artifact_2 = extract(data_integration, DataObject.SENTIMENT)

    assert table_artifact_1.name() == "%s query artifact" % data_integration.name()
    assert table_artifact_2.name() == "%s query artifact" % data_integration.name()

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

    # Check that the names were properly deduplicated at publish time.
    flow = publish_flow_test(client, artifacts=[fn_artifact], engine=engine, name=flow_name())
    flow_run = flow.latest()

    # They both have the same data, but the order shouldn't matter.
    assert flow_run.artifact(table_artifact_1.name()).get().equals(table_artifact_1.get())
    assert flow_run.artifact(table_artifact_1.name() + " (1)").get().equals(table_artifact_1.get())


def test_extract_with_explicit_name_collision(client, data_integration, engine, flow_name):
    # In the case where an explicit op name is supplied twice, we will deduplicate at publish time.
    table_artifact_1 = extract(data_integration, DataObject.SENTIMENT, op_name="sql query")
    assert table_artifact_1.name() == "sql query artifact"

    table_artifact_2 = extract(data_integration, DataObject.SENTIMENT, op_name="sql query")
    assert table_artifact_2.name() == "sql query artifact"

    # Check that the old operator still exists and works.
    table_1 = table_artifact_1.get()
    table_2 = table_artifact_2.get()
    assert table_1.equals(table_2)
    assert list(table_1) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
    ]
    assert table_1.shape[0] == 100

    flow = publish_flow_test(
        client, artifacts=[table_artifact_1, table_artifact_2], engine=engine, name=flow_name()
    )
    flow_run = flow.latest()
    assert flow_run.artifact("sql query artifact").get().equals(table_1)
    assert flow_run.artifact("sql query artifact (1)").get().equals(table_1)


def test_extract_with_custom_artifact(client, data_integration, engine, flow_name):
    output = extract(data_integration, DataObject.SENTIMENT, output_name="hotel reviews")
    assert output.name() == "hotel reviews"

    flow = publish_flow_test(client, artifacts=output, engine=engine, name=flow_name())
    assert flow.latest().artifact("hotel reviews").get().equals(output.get())

    # We can name another output artifact the same, but we can't publish the two together.
    output2 = extract(data_integration, DataObject.SENTIMENT, output_name="hotel reviews")
    with pytest.raises(
        InvalidUserActionException,
        match="Unable to publish flow. You are attempting to publish multiple artifacts explicitly named",
    ):
        client.publish_flow("Test", artifacts=[output, output2], engine=engine)
