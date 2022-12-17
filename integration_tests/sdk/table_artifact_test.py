import math
import time

import pandas as pd
from constants import SENTIMENT_SQL_QUERY, WINE_SQL_QUERY

from aqueduct import op


def test_great_expectations_check(client, data_integration):
    table = data_integration.sql(query=WINE_SQL_QUERY)
    ge_check = table.validate_with_expectation(
        "expect_column_values_to_be_unique", {"column": "fixed_acidity"}
    )

    assert ge_check.get() == False

    ge_check = table.validate_with_expectation(
        "expect_column_values_to_not_be_null", {"column": "fixed_acidity"}
    )

    assert ge_check.get() == True


@op
def corrupt_table_data(table: pd.DataFrame) -> pd.DataFrame:
    index_list = table.index.values.tolist()
    index_list.append(index_list[-1] + 1)
    return table.reindex(index_list)


SLEEP_TIME = 1.1


@op
def timed_function(table: pd.DataFrame) -> pd.DataFrame:
    time.sleep(SLEEP_TIME)
    return table


@op
def mem_intensive_function(table: pd.DataFrame) -> pd.DataFrame:
    a = [0] * 1000
    b = a * 100
    _ = b * 100
    return table


def test_system_runtime_metric(client, data_integration):
    table = data_integration.sql(query=SENTIMENT_SQL_QUERY)
    timed_table = timed_function(table)

    runtime_metric = timed_table.system_metric("runtime")
    runtime = runtime_metric.get()
    assert runtime > SLEEP_TIME


def test_system_max_memory_metric(client, data_integration):
    table = data_integration.sql(query=SENTIMENT_SQL_QUERY)
    timed_table = mem_intensive_function(table)

    max_mem_metric = timed_table.system_metric("max_memory")
    max_mem = max_mem_metric.get()
    assert max_mem > 10


def test_number_of_missing_values(client, data_integration):
    table = data_integration.sql(query=SENTIMENT_SQL_QUERY)
    missing_metric = table.number_of_missing_values(column_id="hotel_name")
    assert missing_metric.get() == 0

    missing_table = corrupt_table_data(table)
    missing_metric = missing_table.number_of_missing_values(column_id="hotel_name")
    assert missing_metric.get() == 1

    missing_metric = missing_table.number_of_missing_values(row_id=100)
    assert missing_metric.get() == 4


def test_number_of_rows(client, data_integration):
    table = data_integration.sql(query=SENTIMENT_SQL_QUERY)
    missing_metric = table.number_of_rows()
    assert missing_metric.get() == 100

    missing_table = corrupt_table_data(table)
    missing_metric = missing_table.number_of_rows()
    assert missing_metric.get() == 101


def test_max(client, data_integration):
    table = data_integration.sql(query=WINE_SQL_QUERY)
    missing_metric = table.max(column_id="fixed_acidity")
    assert math.isclose(missing_metric.get(), 15.8999, rel_tol=1e-3)

    missing_metric = table.max(column_id="total_sulfur_dioxide")
    assert math.isclose(missing_metric.get(), 440, rel_tol=1e-3)


def test_min(client, data_integration):
    table = data_integration.sql(query=WINE_SQL_QUERY)
    missing_metric = table.min(column_id="fixed_acidity")
    assert math.isclose(missing_metric.get(), 3.7999, rel_tol=1e-3)

    missing_metric = table.min(column_id="total_sulfur_dioxide")
    assert math.isclose(missing_metric.get(), 6, rel_tol=1e-3)


def test_mean(client, data_integration):
    table = data_integration.sql(query=WINE_SQL_QUERY)
    missing_metric = table.mean(column_id="fixed_acidity")
    assert math.isclose(missing_metric.get(), 7.2153, rel_tol=1e-3)

    missing_metric = table.mean(column_id="total_sulfur_dioxide")
    assert math.isclose(missing_metric.get(), 115.7445, rel_tol=1e-3)


def test_std(client, data_integration):
    table = data_integration.sql(query=WINE_SQL_QUERY)
    missing_metric = table.std(column_id="fixed_acidity")
    assert math.isclose(missing_metric.get(), 1.2964, rel_tol=1e-3)

    missing_metric = table.std(column_id="total_sulfur_dioxide")
    assert math.isclose(missing_metric.get(), 56.5218, rel_tol=1e-3)


def test_head_standard(client, data_integration):
    table = data_integration.sql(query=SENTIMENT_SQL_QUERY)
    assert table.get().shape[0] == 100

    table_head = table.head()
    assert table_head.shape[0] == 5
