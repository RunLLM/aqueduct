import pytest
from aqueduct import LoadUpdateMode
from aqueduct.error import InvalidIntegrationException

from constants import SENTIMENT_SQL_QUERY
from utils import (
    get_integration_name,
    run_sentiment_model,
    generate_table_name,
)


def test_sql_integration_load_table(sp_client):
    db = sp_client.integration(name=get_integration_name())
    df = db.table(name="hotel_reviews")
    assert len(df) == 100
    assert list(df) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
    ]


def test_invalid_source_integration(sp_client):
    with pytest.raises(InvalidIntegrationException):
        sp_client.integration(name="wrong integration name")


def test_invalid_destination_integration(sp_client):
    db = sp_client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    output_artifact = run_sentiment_model(sql_artifact)

    with pytest.raises(InvalidIntegrationException):
        db._metadata.name = "bad name"
        output_artifact.save(
            config=db.config(table=generate_table_name(), update_mode=LoadUpdateMode.REPLACE)
        )
