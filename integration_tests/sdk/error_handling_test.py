import io
from contextlib import redirect_stdout

import pytest
from aqueduct.error import AqueductError
from utils import get_integration_name

import aqueduct


@aqueduct.op()
def bad_op(df):
    5 / 0
    return df


GOOD_QUERY = "SELECT * FROM hotel_reviews"
BAD_QUERY = "SELEC * FROM sdafawefa"

# These tips should match executor code so that we can verify the correct error is generated.
TIP_EXTRACT = "We couldn't execute the provided query. Please double check your query is correct."
TIP_OP_EXECUTION = "Error executing operator. Please refer to the stack trace for fix."


def test_handle_relational_query_error(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=BAD_QUERY)
    stdout_log = io.StringIO()
    with redirect_stdout(stdout_log), pytest.raises(AqueductError):
        sql_artifact.get()

    stdout_log.seek(0)
    stdout_contents = stdout_log.read()
    assert TIP_EXTRACT in stdout_contents


def test_handle_bad_op_error(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=GOOD_QUERY)
    processed_artifact = bad_op(sql_artifact)
    stdout_log = io.StringIO()
    with redirect_stdout(stdout_log), pytest.raises(AqueductError):
        processed_artifact.get()

    stdout_log.seek(0)
    stdout_contents = stdout_log.read()
    assert TIP_OP_EXECUTION in stdout_contents
