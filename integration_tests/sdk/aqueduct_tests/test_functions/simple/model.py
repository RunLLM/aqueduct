from aqueduct import op


@op()
def dummy_sentiment_model(df):
    df["positivity"] = 123
    return df


def dummy_sentiment_model_function(df):
    df["positivity"] = 123
    return df


@op
def dummy_sentiment_model_multiple_input(df1, df2):
    df1["positivity"] = 123
    df1["positivity_2"] = 456
    return df1


@op()
def dummy_model(df):
    df["newcol"] = "value"
    return df


@op()
def dummy_model_2(df):
    df["newcol_2"] = "value"
    return df
