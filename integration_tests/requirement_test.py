import pandas as pd
import pytest

from constants import SENTIMENT_SQL_QUERY, WINE_SQL_QUERY
from utils import get_integration_name
from aqueduct import op

@op(reqs="~/random.txt")
def invalid_path_operator(table: pd.DataFrame) -> pd.DataFrame:
    return table

@op(reqs="/home/ubuntu/aqueduct/integration_tests/requirement_test.py")
def valid_path_operator(table: pd.DataFrame) -> pd.DataFrame:
    return table

def test_invalid_path_operator(client):
    db = client.integration(name=get_integration_name())
    table = db.sql(query=SENTIMENT_SQL_QUERY)
    with pytest.raises(FileNotFoundError):
        invalid_path_table = invalid_path_operator(table)

def test_valid_path_operator(client):
    db = client.integration(name=get_integration_name())
    table = db.sql(query=SENTIMENT_SQL_QUERY)
    valid_path_table = invalid_path_operator(table)
    assert valid_path_operator.shape[0] == 100
