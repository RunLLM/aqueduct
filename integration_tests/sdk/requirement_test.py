import pandas as pd
import pytest
from aqueduct.error import AqueductError

from constants import SENTIMENT_SQL_QUERY
from utils import get_integration_name
from aqueduct import op

INVALID_REQUIREMENTS_PATH = "~/random.txt"
VALID_REQUIREMENTS_PATH = "requirements/requirements.txt"


def check_if_transformers_exist():
    try:
        import transformers
    except ImportError:
        return False
    return True


@op(reqs_path=INVALID_REQUIREMENTS_PATH)
def invalid_valid_path_operator(table: pd.DataFrame) -> pd.DataFrame:
    return table


@op(reqs_path=VALID_REQUIREMENTS_PATH)
def valid_sentiment_prediction(reviews: pd.DataFrame) -> pd.DataFrame:
    import transformers

    model = transformers.pipeline("sentiment-analysis")
    return reviews.join(pd.DataFrame(model(list(reviews["review"]))))


@op()
def default_sentiment_prediction(reviews: pd.DataFrame) -> pd.DataFrame:
    import transformers

    model = transformers.pipeline("sentiment-analysis")
    return reviews.join(pd.DataFrame(model(list(reviews["review"]))))


def test_invalid_path_operator(client):
    db = client.integration(name=get_integration_name())
    table = db.sql(query=SENTIMENT_SQL_QUERY)
    with pytest.raises(FileNotFoundError):
        invalid_path_table = invalid_valid_path_operator(table)


@pytest.mark.skipif(
    condition=check_if_transformers_exist(),
    reason="the transformers package already exists so the error can't be triggered.",
)
def test_default_path_operator(client):
    db = client.integration(name=get_integration_name())
    table = db.sql(query=SENTIMENT_SQL_QUERY)
    default_path_table = default_sentiment_prediction(table)
    with pytest.raises(AqueductError):
        default_path_table.get()


@pytest.mark.last
def test_valid_path_operator_with_requirement(client):
    db = client.integration(name=get_integration_name())
    table = db.sql(query=SENTIMENT_SQL_QUERY)
    valid_path_table = valid_sentiment_prediction(table)
    assert valid_path_table.get().shape[0] == 100
