from aqueduct.decorator import to_operator
from data_objects import DataObject
from test_function import dummy_sentiment_model_function
from utils import extract

from aqueduct import op


def test_to_operator_local_function(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    @op
    def dummy_sentiment_model(df):
        df["positivity"] = 123
        return df

    def dummy_sentiment_model_func(df):
        df["positivity"] = 123
        return df

    output_artifact_from_decorator = dummy_sentiment_model(table_artifact)
    df_normal = output_artifact_from_decorator.get()
    output_artifact_from_to_operator = to_operator(dummy_sentiment_model_func)(table_artifact)
    df_func = output_artifact_from_to_operator.get()

    assert df_normal["positivity"].equals(df_func["positivity"])


def test_to_operator_imported_function(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    @op(file_dependencies=["test_function.py"])
    def decorated_func(df):
        df = dummy_sentiment_model_function(df)
        return df

    df_decorate = decorated_func(table_artifact).get()
    df_function = to_operator(
        dummy_sentiment_model_function, file_dependencies=["test_function.py"]
    )(table_artifact).get()

    assert df_decorate["positivity"].equals(df_function["positivity"])
