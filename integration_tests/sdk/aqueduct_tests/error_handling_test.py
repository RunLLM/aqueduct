import pytest

import aqueduct
from aqueduct import op
from aqueduct.error import AqueductError, InvalidUserArgumentException

from ..shared.data_objects import DataObject
from .extract import extract


@aqueduct.op()
def bad_op(df):
    5 / 0
    return df


@aqueduct.op(num_outputs=2)
def bad_op_multiple_outputs(df):
    return bad_op(df)


# These tips should match executor code so that we can verify the correct error is generated.
TIP_OP_EXECUTION = "Error executing operator. Please refer to the stack trace for fix."


def test_handle_bad_op_error(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    with pytest.raises(AqueductError, match=TIP_OP_EXECUTION):
        output_artifact = bad_op(table_artifact)
        output_artifact.get()


def test_handle_bad_op_with_multiple_outputs(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    with pytest.raises(AqueductError, match=TIP_OP_EXECUTION):
        output_artifact = bad_op_multiple_outputs(table_artifact)
        output_artifact.get()


def test_file_dependencies_invalid(client):
    with pytest.raises(InvalidUserArgumentException):

        @op(file_dependencies=123)
        def wrong_file_dependencies_type(table):
            return table
