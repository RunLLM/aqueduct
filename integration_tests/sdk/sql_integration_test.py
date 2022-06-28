import pytest
from aqueduct import LoadUpdateMode, metric
from aqueduct.error import InvalidIntegrationException, InvalidUserArgumentException

from constants import SENTIMENT_SQL_QUERY
from utils import (
    get_integration_name,
    run_sentiment_model,
    generate_table_name,
    run_flow_test,
)


def test_sql_integration_load_table(client):
    db = client.integration(name=get_integration_name())
    df = db.table(name="hotel_reviews")
    assert len(df) == 100
    assert list(df) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
    ]


def test_invalid_source_integration(client):
    with pytest.raises(InvalidIntegrationException):
        client.integration(name="wrong integration name")


def test_invalid_destination_integration(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    output_artifact = run_sentiment_model(sql_artifact)

    with pytest.raises(InvalidIntegrationException):
        db._metadata.name = "bad name"
        output_artifact.save(
            config=db.config(table=generate_table_name(), update_mode=LoadUpdateMode.REPLACE)
        )


def test_sql_today_tag(client):
    db = client.integration(name=get_integration_name())
    sql_artifact_today = db.sql(query="select * from hotel_reviews where review_date = {{today}}")
    assert sql_artifact_today.get().empty
    sql_artifact_not_today = db.sql(
        query="select * from hotel_reviews where review_date < {{today}}"
    )
    assert len(sql_artifact_not_today.get()) == 100


def test_sql_query_with_parameter(client):
    db = client.integration(name=get_integration_name())

    # Missing parameters.
    with pytest.raises(InvalidUserArgumentException):
        _ = db.sql(query="select * from {{missing_parameter}}")

    # The parameter is not a string type.
    _ = client.create_param("table_name", default=1234)
    with pytest.raises(InvalidUserArgumentException):
        _ = db.sql(query="select * from {{ table_name }}")

    client.create_param("table_name", default="hotel_reviews")
    sql_artifact = db.sql(query="select * from {{ table_name }}")

    expected_sql_artifact = db.sql(query="select * from hotel_reviews")
    assert sql_artifact.get().equals(expected_sql_artifact.get())
    expected_sql_artifact = db.sql(query="select * from customer_activity")
    assert sql_artifact.get(parameters={"table_name": "customer_activity"}).equals(
        expected_sql_artifact.get()
    )

    # Trigger the parameter with invalid values.
    with pytest.raises(InvalidUserArgumentException):
        _ = sql_artifact.get(parameters={"table_name": ["this is the incorrect type"]})
    with pytest.raises(InvalidUserArgumentException):
        _ = sql_artifact.get(parameters={"non-existant parameter": "blah"})


def test_sql_query_with_multiple_parameters(client):
    db = client.integration(name=get_integration_name())

    _ = client.create_param("table_name", default="hotel_reviews")
    nationality = client.create_param("reviewer_nationality", default="United Kingdom")
    sql_artifact = db.sql(
        query="select * from {{ table_name }} where reviewer_nationality='{{ reviewer_nationality }}' and review_date < {{ today}}"
    )
    expected_sql_artifact = db.sql(
        "select * from hotel_reviews where reviewer_nationality='United Kingdom' and review_date < {{today}}"
    )
    assert sql_artifact.get().equals(expected_sql_artifact.get())
    expected_sql_artifact = db.sql(
        "select * from hotel_reviews where reviewer_nationality='Australia' and review_date < {{today}}"
    )
    assert sql_artifact.get(parameters={"reviewer_nationality": "Australia"}).equals(
        expected_sql_artifact.get()
    )

    # Use the parameters in another operator.
    @metric
    def noop(sql_output, param):
        return len(param)

    result = noop(sql_artifact, nationality)
    assert result.get() == len(nationality.get())
    assert result.get(parameters={"reviewer_nationality": "Australia"}) == len("Australia")

    run_flow_test(client, artifacts=[result])
