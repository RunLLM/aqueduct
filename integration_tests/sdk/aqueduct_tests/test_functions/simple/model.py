import pandas as pd

from aqueduct import op


# In order to use these functions with spark compute engines, we add a clause
# with equivalent pyspark code. the `lit` function creates a full column of the
# given value. `lit(123)` automatically is casted to a double, so we case back
# to an integer type.
@op()
def dummy_sentiment_model(df):
    if isinstance(df, pd.DataFrame):
        df["positivity"] = 123
    else:
        from pyspark.sql.functions import lit

        df = df.withColumn("POSITIVITY", lit(123).cast("integer"))

    return df


def dummy_sentiment_model_function(df):
    if isinstance(df, pd.DataFrame):
        df["positivity"] = 123
    else:
        from pyspark.sql.functions import lit

        df = df.withColumn("POSITIVITY", lit(123).cast("integer"))

    return df


@op
def dummy_sentiment_model_multiple_input(df1, df2):
    if isinstance(df1, pd.DataFrame):
        df1["positivity"] = 123
        df1["positivity_2"] = 456
    else:
        from pyspark.sql.functions import lit

        df1 = df1.withColumn("POSITIVITY", lit(123).cast("integer"))
        df1 = df1.withColumn("POSITIVITY_2", lit(456).cast("integer"))

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
