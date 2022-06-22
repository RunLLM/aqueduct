import pandas as pd
import pytest 
import transformers

from constants import SENTIMENT_SQL_QUERY
from utils import get_integration_name
from aqueduct import op

@op(reqs_path="~/random.txt")
def error_valid_path_operator(table: pd.DataFrame) -> pd.DataFrame:
    return table

@op(reqs_path="requirements/requirements.txt")
def valid_sentiment_prediction(reviews: pd.DataFrame) -> pd.DataFrame:
    model = transformers.pipeline("sentiment-analysis")
    return reviews.join(pd.DataFrame(model(list(reviews['review']))))

@op()
def invalid_sentiment_prediction(reviews: pd.DataFrame) -> pd.DataFrame:
    model = transformers.pipeline("sentiment-analysis")
    return reviews.join(pd.DataFrame(model(list(reviews['review']))))

def test_invalid_path_operator(client):
    db = client.integration(name=get_integration_name())
    table = db.sql(query=SENTIMENT_SQL_QUERY)
    with pytest.raises(FileNotFoundError):
        invalid_path_table = error_valid_path_operator(table)

def test_valid_path_operator_with_requirement(client):
    db = client.integration(name=get_integration_name())
    table = db.sql(query=SENTIMENT_SQL_QUERY)
    valid_path_table = valid_sentiment_prediction(table)
    assert valid_path_table.get().shape[0] == 100
