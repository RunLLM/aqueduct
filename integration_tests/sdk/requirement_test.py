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
def sentiment_prediction_with_invalid_reqs_path(table: pd.DataFrame) -> pd.DataFrame:
    return table


@op(reqs_path=VALID_REQUIREMENTS_PATH)
def sentiment_prediction_with_valid_reqs_path(reviews: pd.DataFrame) -> pd.DataFrame:
    import transformers

    model = transformers.pipeline("sentiment-analysis")
    return reviews.join(pd.DataFrame(model(list(reviews["review"]))))


@op()
def sentiment_prediction_without_reqs_path(reviews: pd.DataFrame) -> pd.DataFrame:
    import transformers

    model = transformers.pipeline("sentiment-analysis")
    return reviews.join(pd.DataFrame(model(list(reviews["review"]))))


def test_invalid_path_operator(client):
    db = client.integration(name=get_integration_name())
    table = db.sql(query=SENTIMENT_SQL_QUERY)
    with pytest.raises(FileNotFoundError):
        invalid_path_table = sentiment_prediction_with_invalid_reqs_path(table)


@pytest.mark.skipif(
    condition=check_if_transformers_exist(),
    reason="the transformers package already exists so the error can't be triggered.",
)
def test_path_operator(client):
    db = client.integration(name=get_integration_name())
    table = db.sql(query=SENTIMENT_SQL_QUERY)
    # test if operator with default path will error out because of missing requirement package
    default_path_table = sentiment_prediction_without_reqs_path(table)
    with pytest.raises(AqueductError):
        default_path_table.get()
    # test if operator with valid requirement path will install the package and return the correct dataframe
    valid_path_table = sentiment_prediction_with_valid_reqs_path(table)
    assert valid_path_table.get().shape[0] == 100
