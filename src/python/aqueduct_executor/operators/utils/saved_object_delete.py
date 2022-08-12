from aqueduct_executor.operators.utils.execution import ExecutionState
from pydantic import BaseModel


class SavedObjectDelete(BaseModel):
    """This contains the result of deleting the saved object."""

    name: str
    exec_state: ExecutionState
