import pytest
from aqueduct.error import InvalidIntegrationException, InvalidUserArgumentException
from data_objects import DataObject
from relational import all_relational_DBs
from test_functions.simple.model import dummy_sentiment_model
from utils import extract, publish_flow_test, save

from aqueduct import metric


@pytest.mark.enable_only_for_data_integration_type(*all_relational_DBs())
def test_sql_integration_load_table(client, data_integration):
    df = data_integration.table(name="hotel_reviews")
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


def test_invalid_destination_integration(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)

    with pytest.raises(InvalidIntegrationException):
        data_integration._metadata.name = "bad name"
        save(data_integration, output_artifact)


@pytest.mark.enable_only_for_data_integration_type(*all_relational_DBs())
def test_sql_today_tag(client, data_integration):
    table_artifact_today = data_integration.sql(
        query="select * from hotel_reviews where review_date = {{today}}"
    )
    assert table_artifact_today.get().empty
    table_artifact_not_today = data_integration.sql(
        query="select * from hotel_reviews where review_date < {{today}}"
    )
    assert len(table_artifact_not_today.get()) == 100


@pytest.mark.enable_only_for_data_integration_type(*all_relational_DBs())
def test_sql_query_with_parameter(client, data_integration):
    # Missing parameters.
    with pytest.raises(InvalidUserArgumentException):
        _ = data_integration.sql(query="select * from {{missing_parameter}}")

    # The parameter is not a string type.
    _ = client.create_param("table_name", default=1234)
    with pytest.raises(InvalidUserArgumentException):
        _ = data_integration.sql(query="select * from {{ table_name }}")

    client.create_param("table_name", default="hotel_reviews")
    table_artifact = data_integration.sql(query="select * from {{ table_name }}")

    expected_table_artifact = data_integration.sql(query="select * from hotel_reviews")
    assert table_artifact.get().equals(expected_table_artifact.get())
    expected_table_artifact = data_integration.sql(query="select * from customer_activity")
    assert table_artifact.get(parameters={"table_name": "customer_activity"}).equals(
        expected_table_artifact.get()
    )

    # Trigger the parameter with invalid values.
    with pytest.raises(InvalidUserArgumentException):
        _ = table_artifact.get(parameters={"table_name": ["this is the incorrect type"]})
    with pytest.raises(InvalidUserArgumentException):
        _ = table_artifact.get(parameters={"non-existant parameter": "blah"})


@pytest.mark.enable_only_for_data_integration_type(*all_relational_DBs())
def test_sql_query_with_multiple_parameters(client, flow_name, data_integration, engine):
    _ = client.create_param("table_name", default="hotel_reviews")
    nationality = client.create_param(
        "reviewer-nationality", default="United Kingdom"
    )  # check that dashes work.
    table_artifact = data_integration.sql(
        query="select * from {{ table_name }} where reviewer_nationality='{{ reviewer-nationality }}' and review_date < {{ today}}"
    )
    expected_table_artifact = data_integration.sql(
        "select * from hotel_reviews where reviewer_nationality='United Kingdom' and review_date < {{today}}"
    )
    assert table_artifact.get().equals(expected_table_artifact.get())
    expected_table_artifact = data_integration.sql(
        "select * from hotel_reviews where reviewer_nationality='Australia' and review_date < {{today}}"
    )
    assert table_artifact.get(parameters={"reviewer-nationality": "Australia"}).equals(
        expected_table_artifact.get()
    )

    # Use the parameters in another operator.
    @metric
    def noop(sql_output, param):
        return len(param)

    result = noop(table_artifact, nationality)
    assert result.get() == len(nationality.get())
    assert result.get(parameters={"reviewer-nationality": "Australia"}) == len("Australia")

    publish_flow_test(client, name=flow_name(), artifacts=[result], engine=engine)


@pytest.mark.enable_only_for_data_integration_type(*all_relational_DBs())
def test_sql_query_user_vs_builtin_precedence(client, data_integration):
    """If a user defines an expansion that collides with a built-in one, the user-defined one should take precedence."""
    table_artifact = data_integration.sql(
        query="select * from hotel_reviews where review_date > {{today}}"
    )
    builtin_result = table_artifact.get()

    datestring = "'2016-01-01'"
    _ = client.create_param("today", datestring)
    table_artifact = data_integration.sql(
        query="select * from hotel_reviews where review_date > {{today}}"
    )
    user_param_result = table_artifact.get()
    assert not builtin_result.equals(user_param_result)

    expected_table_artifact = data_integration.sql(
        query="select * from hotel_reviews where review_date > %s" % datestring
    )
    assert user_param_result.equals(expected_table_artifact.get())


@pytest.mark.enable_only_for_data_integration_type(*all_relational_DBs())
def test_chained_sql_query(client):
    client.create_param("nationality", default=" United Kingdom ")
    warehouse = client.integration(name="aqueduct_demo")
    reviews = warehouse.sql(
        [
            """
        SELECT * FROM hotel_reviews
    """,
            " SELECT review, review_date from $ where reviewer_nationality ='{{nationality}}'",
            " SELECT review from $",
        ]
    ).get()
    expected_artf = warehouse.sql(
        "SELECT review FROM hotel_reviews WHERE reviewer_nationality=' United Kingdom '"
    ).get()
    assert reviews.equals(expected_artf)
