from typing import Optional

from aqueduct.constants.enums import ExecutionStatus
from aqueduct.error import ResourceConnectionInProgress, ResourceFailedToConnect
from aqueduct.models.execution_state import ExecutionState


def validate_resource_is_connected(name: str, exec_state: Optional[ExecutionState]) -> None:
    """Method used to determine if this resource was successfully connected to or not.
    If not successfully connected (or pending), we will raise an Exception.
    """
    # TODO(ENG-2813): Remove the assumption that a missing `exec_state` means success.
    if exec_state is None or exec_state.status == ExecutionStatus.SUCCEEDED:
        return

    if exec_state.status == ExecutionStatus.FAILED:
        assert exec_state.error is not None
        raise ResourceFailedToConnect(
            "Cannot use resource %s because it has not been successfully connected to: "
            "%s\n%s\n\n Please see the /resources page on the UI for more details."
            % (
                name,
                exec_state.error.tip,
                exec_state.error.context,
            )
        )
    else:
        # The assumption is that we are in the running state here.
        raise ResourceConnectionInProgress(
            "Cannot use resource %s because it is still in the process of connecting."
            "Please see the /resources page on the UI for more details." % name
        )
