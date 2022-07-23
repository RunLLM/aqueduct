from aqueduct.decorator import to_operator
from constants import SENTIMENT_SQL_QUERY
from aqueduct import op
from utils import get_integration_name, run_sentiment_model

def test_to_operator(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)
    @op
    def dummy_sentiment_model(df):
        df["positivity"] = 123
        return df
    
    def dummy_sentiment_model_func(df):
        df["positivity"] = 123
        return df

    output_artifact = dummy_sentiment_model(sql_artifact)
    df_normal = output_artifact.get()
    output_artifact_func = to_operator(dummy_sentiment_model_func)(sql_artifact)
    df_func = output_artifact_func.get()

    assert(df_normal["positivity"].equals(df_func["positivity"]))
