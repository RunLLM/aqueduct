import pandas as pd

from aqueduct import op


@op()
def dummy_sentiment_model(df):
    if isinstance(df, pd.DataFrame):
        df["positivity"] = 123
    else:
        from pyspark.sql.functions import lit

        df = df.withColumn("POSITIVITY", lit(123.0))

    return df


def dummy_sentiment_model_function(df):
    if isinstance(df, pd.DataFrame):
        df["positivity"] = 123
    else:
        from pyspark.sql.functions import lit

        df = df.withColumn("POSITIVITY", lit(123.0))

    return df


@op
def dummy_sentiment_model_multiple_input(df1, df2):
    if isinstance(df1, pd.DataFrame):
        df1["positivity"] = 123
        df1["positivity_2"] = 456
    else:
        from pyspark.sql.functions import lit

        df1 = df1.withColumn("POSITIVITY", lit(123.0))
        df1 = df1.withColumn("POSITIVITY_2", lit(456.0))

    return df1


@op()
def dummy_model(df):
    if isinstance(df, pd.DataFrame):
        df["newcol"] = "value"
    else:
        from pyspark.sql.functions import lit

        df = df.withColumn("NEWCOL", lit("value"))

    return df


@op()
def dummy_model_2(df):
    if isinstance(df, pd.DataFrame):
        df["newcol_2"] = "value"
    else:
        from pyspark.sql.functions import lit

        df = df.withColumn("NEWCOL_2", lit("value"))

    return df
