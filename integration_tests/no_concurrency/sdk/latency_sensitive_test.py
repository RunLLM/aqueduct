from utils import SENTIMENT_SQL_QUERY, get_integration_name

from aqueduct import op


def test_preview_artifact_caching(client):
    db = client.resource(name=get_integration_name())
    sql_artifact = db.sql(query=SENTIMENT_SQL_QUERY)

    @op
    def slow_fn(df):
        time.sleep(5)
        return df

    @op
    def noop(df):
        return df

    # Check that the first run will take a while, but the second run will happen much faster.
    import time

    start = time.time()
    slow_output = slow_fn(sql_artifact)
    first_duration = time.time() - start
    assert first_duration > 5

    start = time.time()
    _ = noop(slow_output)
    assert time.time() - start < first_duration
