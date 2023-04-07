from aqueduct import op
import pandas as pd


@op()
def dummy_sentiment_model(df):
    if isinstance(df, pd.DataFrame):
        df["positivity"] = 123
    else:
        from pyspark.sql.functions import lit
        df = df.withColumn("positivity", lit(123))

    return df


def dummy_sentiment_model_function(df):
    if isinstance(df, pd.DataFrame):
        df["positivity"] = 123
    else:
        from pyspark.sql.functions import lit
        df = df.withColumn("positivity", lit(123))

    return df


@op
def dummy_sentiment_model_multiple_input(df1, df2):
    if isinstance(df, pd.DataFrame):
        df1["positivity"] = 123
        df1["positivity_2"] = 456
    else:
        from pyspark.sql.functions import lit
        df1 = df1.withColumn("positivity", lit(123))
        df1 = df1.withColumn("positivity_2", lit(456))

    return df1


@op()
def dummy_model(df):
    if isinstance(df, pd.DataFrame):
        df["newcol"] = "value"
    else:
        from pyspark.sql.functions import lit
        df = df.withColumn("newcol", lit("value"))

    return df


@op()
def dummy_model_2(df):
    if isinstance(df, pd.DataFrame):
        df["newcol_2"] = "value"
    else:
        from pyspark.sql.functions import lit
        df = df.withColumn("newcol_2", lit("value"))

    return df
