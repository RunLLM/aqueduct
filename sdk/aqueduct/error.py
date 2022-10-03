class Error(Exception):
    def __init__(self, message: str):
        """Exception raised for different kinds of errors.
        Attributes:
                message: explanation of the error
        """
        self.message = message


# Exception raised for errors occured when certain inputs are missing or contain
# invalid values. Also thrown on any errors in the user's queries or function code.
class InvalidRequestError(Error):
    pass


# Exception raised for errors occured when the Aqueduct system fails to process
# certain inputs.
class UnprocessableEntityError(Error):
    pass


# Exception raised when an internal error occured within the Aqueduct system.
class InternalServerError(Error):
    def __init__(self, message: str):
        self.message = (
            "An internal server error occured when processing the "
            "request. This is likely not your fault. Please contact "
            "us at support@aqueducthq.com with your Aqueduct account "
            "email and a brief description of what you were trying to "
            "do when the error occurred, so that we can further "
            "assist you. Apologies for the inconvenience!"
        )


# Exception raised when the requested resource is not found.
class ResourceNotFoundError(Error):
    pass


# Exception raised when the requested metadata field is invalid.
class InvalidMetadataError(Error):
    pass


# Exception raised when the type of an artifact does not meet the expectation.
class InvalidArtifactTypeException(Error):
    pass


# A catch-all exception raised for all other errors.
class AqueductError(Error):
    pass


# An exception that indicates something is wrong within the system, not the user.
class InternalAqueductError(Error):
    pass


"""
BELOW: Errors that are raised by the Python SDK code, not the backend.
"""


# Exception raised when trying to create flow without connecting to integration
class NoConnectedIntegrationsException(Error):
    pass


# Exception raised when referencing an integration that doesn't exist or the wrong
# integration type.
class InvalidIntegrationException(Error):
    pass


# Exception raised when checking a flow that doesn't write to any destination store.
class NoDestinationIntegrationException(Error):
    pass


# Exception raised when the user misconfigures something when running a function.
class InvalidFunctionException(Error):
    pass


# Exception raised when using invalid cron string
class InvalidCronStringException(Error):
    pass


# Exception raised when an inappropriate user action is attempted.
class InvalidUserActionException(Error):
    pass


# Exception raised when an inappropriate argument has been supplied by the user.
class InvalidUserArgumentException(Error):
    pass


# Exception raised when user attempts to use an invalid file name as a file dependency.
class ReservedFileNameException(Error):
    pass


class ArtifactNotFoundException(Error):
    pass


class ArtifactNeverComputedException(Error):
    pass


# Exception raised when user tries to use a file defined outside of provided function
# directory as a dependency.
class InvalidDependencyFilePath(Error):
    pass


# Exception raised when a github query is not valid.
class InvalidGithubQueryError(Error):
    pass


# Exception raised when client fails to validate the server.
class ClientValidationError(Error):
    pass
