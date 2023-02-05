import pandas as pd
import pytest
from aqueduct.error import InvalidUserActionException
from aqueduct.integrations.sql_integration import RelationalDBIntegration

from aqueduct import LoadUpdateMode, op

from ..shared.naming import generate_table_name
from .save import save


@pytest.fixture(autouse=True)
def assert_data_integration_is_relational(client, data_integration):
    assert isinstance(data_integration, RelationalDBIntegration)


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
