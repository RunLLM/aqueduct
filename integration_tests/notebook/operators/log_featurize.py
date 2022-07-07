from aqueduct import op
import numpy as np
import os
import pandas as pd


@op(file_dependencies=["./data"])
def log_featurize(cust: pd.DataFrame) -> pd.DataFrame:
    if not os.path.exists("data"):
        raise Exception("Data does not exist!")
    features = cust.copy()
    skip_cols = ["cust_id", "using_deep_learning", "using_dbt"]
    for col in features.columns.difference(skip_cols):
        features["log_" + col] = np.log(features[col] + 1.0)
    return features.drop(columns="cust_id")
