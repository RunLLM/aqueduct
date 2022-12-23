import pytest
from aqueduct.error import AqueductError, InvalidUserArgumentException
from data_objects import DataObject
from relational import all_relational_DBs
from utils import extract

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


@pytest.mark.enable_only_for_data_integration_type(*all_relational_DBs())
def test_handle_relational_query_error(client, data_integration):
    try:
        _ = data_integration.sql(query=BAD_QUERY)
    except AqueductError as e:
        assert TIP_EXTRACT in e.message


def test_handle_bad_op_error(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    try:
        bad_op(table_artifact)
    except AqueductError as e:
        assert TIP_OP_EXECUTION in e.message


def test_file_dependencies_invalid(client):
    with pytest.raises(InvalidUserArgumentException):

        @op(file_dependencies=123)
        def wrong_file_dependencies_type(table):
            return table
