from aqueduct.decorator import to_operator
from constants import SENTIMENT_SQL_QUERY
from aqueduct import op
from utils import get_integration_name, run_sentiment_model
from test_function import dummy_sentiment_model_function


def test_to_operator_local_function(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)

    @op
    def dummy_sentiment_model(df):
        df["positivity"] = 123
        return df

    def dummy_sentiment_model_func(df):
        df["positivity"] = 123
        return df

    output_artifact_from_decorator = dummy_sentiment_model(sql_artifact)
    df_normal = output_artifact_from_decorator.get()
    output_artifact_from_to_operator = to_operator(dummy_sentiment_model_func)(sql_artifact)
    df_func = output_artifact_from_to_operator.get()

    assert df_normal["positivity"].equals(df_func["positivity"])


def test_to_operator_imported_function(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)

    @op(file_dependencies=["function.py"])
    def decorated_func(df):
        df = dummy_sentiment_model_function(df)
        return df

    df_decorate = decorated_func(sql_artifact).get()
    df_function = to_operator(
        dummy_sentiment_model_function, file_dependencies=["test_function.py"]
    )(sql_artifact).get()

    assert df_decorate["positivity"].equals(df_function["positivity"])
