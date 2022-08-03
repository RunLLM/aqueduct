import io
from contextlib import redirect_stdout

import pytest
from aqueduct.error import AqueductError, InvalidUserArgumentException
from utils import get_integration_name

import aqueduct
from aqueduct import op


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

    try:
        sql_artifact.get()
    except AqueductError as e:
        assert TIP_EXTRACT in e.message


def test_handle_bad_op_error(client):
    db = client.integration(name=get_integration_name())
    sql_artifact = db.sql(query=GOOD_QUERY)
    processed_artifact = bad_op(sql_artifact)

    try:
        processed_artifact.get()
    except AqueductError as e:
        assert TIP_OP_EXECUTION in e.message


def test_file_dependencies_invalid(client):
    with pytest.raises(InvalidUserArgumentException):

        @op(file_dependencies=123)
        def wrong_file_dependencies_type(table):
            return table
