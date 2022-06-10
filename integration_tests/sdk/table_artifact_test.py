import pandas as pd
import math

from constants import SENTIMENT_SQL_QUERY, WINE_SQL_QUERY
from utils import get_integration_name
from aqueduct import op


@op
def corrupt_table_data(table: pd.DataFrame) -> pd.DataFrame:
    index_list = table.index.values.tolist()
    index_list.append(index_list[-1] + 1)
    return table.reindex(index_list)


def test_number_of_missing_values(client):
    db = client.integration(name=get_integration_name())
    table = db.sql(query=SENTIMENT_SQL_QUERY)
    missing_metric = table.number_of_missing_values(column_id="hotel_name")
    assert missing_metric.get() == 0

    missing_table = corrupt_table_data(table)
    missing_metric = missing_table.number_of_missing_values(column_id="hotel_name")
    assert missing_metric.get() == 1

    missing_metric = missing_table.number_of_missing_values(row_id=100)
    assert missing_metric.get() == 4


def test_number_of_rows(client):
    db = client.integration(name=get_integration_name())
    table = db.sql(query=SENTIMENT_SQL_QUERY)
    missing_metric = table.number_of_rows()
    assert missing_metric.get() == 100

    missing_table = corrupt_table_data(table)
    missing_metric = missing_table.number_of_rows()
    assert missing_metric.get() == 101


def test_max(client):
    db = client.integration(name=get_integration_name())
    table = db.sql(query=WINE_SQL_QUERY)
    missing_metric = table.max(column_id="fixed_acidity")
    assert math.isclose(missing_metric.get(), 15.8999, rel_tol=1e-3)

    missing_metric = table.max(column_id="total_sulfur_dioxide")
    assert math.isclose(missing_metric.get(), 440, rel_tol=1e-3)


def test_min(client):
    db = client.integration(name=get_integration_name())
    table = db.sql(query=WINE_SQL_QUERY)
    missing_metric = table.min(column_id="fixed_acidity")
    assert math.isclose(missing_metric.get(), 3.7999, rel_tol=1e-3)

    missing_metric = table.min(column_id="total_sulfur_dioxide")
    assert math.isclose(missing_metric.get(), 6, rel_tol=1e-3)


def test_mean(client):
    db = client.integration(name=get_integration_name())
    table = db.sql(query=WINE_SQL_QUERY)
    missing_metric = table.mean(column_id="fixed_acidity")
    assert math.isclose(missing_metric.get(), 7.2153, rel_tol=1e-3)

    missing_metric = table.mean(column_id="total_sulfur_dioxide")
    assert math.isclose(missing_metric.get(), 115.7445, rel_tol=1e-3)


def test_std(client):
    db = client.integration(name=get_integration_name())
    table = db.sql(query=WINE_SQL_QUERY)
    missing_metric = table.std(column_id="fixed_acidity")
    assert math.isclose(missing_metric.get(), 1.2964, rel_tol=1e-3)

    missing_metric = table.std(column_id="total_sulfur_dioxide")
    assert math.isclose(missing_metric.get(), 56.5218, rel_tol=1e-3)
