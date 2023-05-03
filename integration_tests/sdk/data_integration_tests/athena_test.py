import pandas as pd
import pytest
from aqueduct.error import InvalidUserActionException, InvalidUserArgumentException
from aqueduct.resources.sql import RelationalDBResource

from aqueduct import LoadUpdateMode, metric, op

from ..shared.naming import generate_table_name
from .save import save


@pytest.fixture(autouse=True)
def assert_data_integration_is_relational(client, data_integration):
    assert isinstance(data_integration, RelationalDBResource)


def test_athena_integration_table_retrieval(client, data_integration):
    df = data_integration.table(name="hotel_reviews")
    assert len(df) == 100
    assert list(df) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
    ]


def test_athena_integration_list_tables(client, data_integration):
    expected_tables = [
        "customers",
        "hotel_reviews",
    ]
    tables = data_integration.list_tables()

    for expected_table in expected_tables:
        assert tables["tablename"].str.contains(expected_table, case=False).sum() > 0


def test_athena_save(client, data_integration):
    @op
    def generate_table():
        return pd.DataFrame()

    with pytest.raises(
        InvalidUserActionException,
        match="Save operation not supported for Athena.",
    ):
        save(data_integration, generate_table(), generate_table_name(), LoadUpdateMode.REPLACE)


def test_athena_query_with_parameter(client, data_integration):
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

    # Trigger the parameter with invalid values.
    with pytest.raises(InvalidUserArgumentException):
        _ = table_artifact.get(parameters={"table_name": ["this is the incorrect type"]})
    with pytest.raises(InvalidUserArgumentException):
        _ = table_artifact.get(parameters={"non-existant parameter": "blah"})


def test_athena_query_with_multiple_parameters(client, flow_manager, data_integration):
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

    flow_manager.publish_flow_test(artifacts=[result])
