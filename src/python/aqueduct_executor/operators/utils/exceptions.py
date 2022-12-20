from aqueduct_executor.operators.utils.enums import FailureType


class MissingConnectorDependencyException(Exception):
    """Exception raised due to the connector integration's dependencies aren't installed."""

    pass


class MissingInputPathsException(Exception):
    """Exception raised due to input data not being supplied."""

    pass
