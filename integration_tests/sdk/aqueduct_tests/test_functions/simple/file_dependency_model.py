import os

import pandas as pd

from aqueduct import op


@op(file_dependencies=["./data"])
def model_with_file_dependency(df):
    if not os.path.exists("data"):
        raise Exception("Data does not exist!")

    if isinstance(df, pd.DataFrame):
        df["newcol"] = 999
    else:
        from pyspark.sql.functions import lit

        df = df.withColumn("NEWCOL", lit(999).cast("integer"))

    return df


@op(file_dependencies=["./model.py"])
def model_with_invalid_dependencies(df):
    if not os.path.exists("data"):
        raise Exception("Data does not exist!")
    return df


@op()
def model_with_missing_file_dependencies(df):
    if not os.path.exists("data"):
        raise Exception("Data does not exist!")
    return df


@op(file_dependencies=["gibberuish"])
def model_with_improper_dependency_path(df):
    return df


@op(file_dependencies=["../sentiment/model.py"])
def model_with_out_of_package_file_dependency(df):
    return df
