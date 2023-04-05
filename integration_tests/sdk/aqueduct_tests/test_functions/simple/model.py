from aqueduct import op


@op()
def dummy_sentiment_model(df):
    columns = df.columns
    return df


def dummy_sentiment_model_function(df):
    columns = df.columns
    return df


@op
def dummy_sentiment_model_multiple_input(df1, df2):
    columns = df1.columns
    return df1


@op()
def dummy_model(df):
    columns = df.columns
    return df


@op()
def dummy_model_2(df):
    columns = df.columns
    return df
