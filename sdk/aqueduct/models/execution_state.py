from typing import Optional

from aqueduct.constants.enums import ExecutionStatus, FailureType
from pydantic import BaseModel


class Logs(BaseModel):
    stdout: str = ""
    stderr: str = ""

    def is_empty(self) -> bool:
        return self.stdout == "" and self.stderr == ""


class Error(BaseModel):
    context: str = ""
    tip: str = ""


class ExecutionState(BaseModel):
    user_logs: Optional[Logs] = None
    error: Optional[Error] = None
    status: ExecutionStatus = ExecutionStatus.UNKNOWN
    failure_type: Optional[FailureType] = None
