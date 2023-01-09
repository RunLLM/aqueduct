from aqueduct import metric


@metric()
def constant_metric(df):
    return 17.5
