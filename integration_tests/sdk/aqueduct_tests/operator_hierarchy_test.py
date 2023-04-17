import pytest

from aqueduct import check, metric, op
from aqueduct.error import InvalidUserActionException

from ..shared.data_objects import DataObject
from .extract import extract


@check()
def produce_check_artifact(args):
    return True


@metric()
def produce_metric_artifact(args):
    return 1.0


@check()
def check_function(args):
    return True


@metric()
def metric_function(args):
    return 1.0


@op()
def regular_function(args):
    return "Hello World"


def test_check_artifact_restriction(client, data_integration):
    """Test that an artifact produced by a check operator cannot be used as an argument to any operator types."""
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    check_artifact = produce_check_artifact(table_artifact)
    with pytest.raises(InvalidUserActionException):
        check_function(check_artifact)
    with pytest.raises(InvalidUserActionException):
        metric_function(check_artifact)
    with pytest.raises(InvalidUserActionException):
        regular_function(check_artifact)


def test_metric_artifact_restriction(client, data_integration):
    """Test that an artifact produced by a metric operator cannot be used as an argument to function operator."""
    table_artifact = extract(data_integration, DataObject.SENTIMENT)

    metric_artifact = produce_metric_artifact(table_artifact)
    with pytest.raises(InvalidUserActionException):
        regular_function(metric_artifact)
