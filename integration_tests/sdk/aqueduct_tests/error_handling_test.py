import pytest
from aqueduct.error import AqueductError, InvalidUserArgumentException

import aqueduct
from aqueduct import op

from ..shared.data_objects import DataObject
from .extract import extract


@aqueduct.op()
def bad_op(df):
    5 / 0
    return df


# These tips should match executor code so that we can verify the correct error is generated.
TIP_OP_EXECUTION = "Error executing operator. Please refer to the stack trace for fix."


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
