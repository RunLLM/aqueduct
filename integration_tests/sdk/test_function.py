from aqueduct import op


def dummy_sentiment_model_function(df):
    df["positivity"] = 123
    return df
