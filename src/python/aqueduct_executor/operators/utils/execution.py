from __future__ import annotations

import io
import sys
import traceback
from contextlib import redirect_stderr, redirect_stdout
from typing import Any, Callable, Optional, Type, TypeVar

from aqueduct_executor.operators.utils.enums import ExecutionStatus, FailureType
from pydantic import BaseModel

_GITHUB_ISSUE_LINK = "https://github.com/aqueducthq/aqueduct/issues/new?assignees=&labels=bug&template=bug_report.md&title=%5BBUG%5D"

TIP_OP_EXECUTION = "Error executing operator. Please refer to the stack trace for fix."
_TIP_CREATE_BUG_REPORT = (
    f"Please create bug report in github: {_GITHUB_ISSUE_LINK} . "
    "We will get back to you as soon as we can."
)
TIP_UNKNOWN_ERROR = f"Sorry, we've run into an unexpected error! {_TIP_CREATE_BUG_REPORT}"
TIP_INTEGRATION_CONNECTION = (
    "We were unable to connect to this integration. "
    "Please check your credentials or contact your integration's provider."
)
TIP_DEMO_CONNECTION = f"We have trouble connecting to demo DB. {_TIP_CREATE_BUG_REPORT}"

TIP_EXTRACT = "We couldn't execute the provided query. Please double check your query is correct."
TIP_LOAD = "We couldn't load to the integration. Please make sure the target exists, or you have the right permission."
TIP_DISCOVER = "We couldn't list items in the integration. Please make sure your credentials have the right permission."

# Assumption: only check operators will use this tip.
TIP_CHECK_DID_NOT_PASS = "The check did not pass (returned False)."

TIP_NOT_NUMERIC = "The computed result is not of type numeric."
TIP_NOT_BOOL = "The computed result is not of type bool."


class Error(BaseModel):
    tip: str = ""  # Information about how the user could fix the error.
    context: str = ""  # More details about the error. Typically a stack trace.


class Logs(BaseModel):
    stdout: str = ""
    stderr: str = ""


class ExecFailureException(Exception):
    """Can be thrown whenever you want to fail the operator.

    It captures all the information needed to mark an ExecutionState as failed, except for user logs.
    Can be translated into an ExecutionState with `ExecutionState.from_exception()`.
    """

    def __init__(
        self,
        failure_type: FailureType,
        tip: str,
        context: str = "",
    ):
        self.failure_type = failure_type
        self.tip = tip
        self.context = context


class ExecutionState(BaseModel):
    """
    The state to track operator execution. In the future, we may extend this
    to track arbitrary execution.

    `status`: the status of execution, one of 'pending', 'succeeded', or 'failed'.
    `user_logs`: the stderr and stdout of 'user' part of execution. Available regardless of status.
    `failure_type`: more detailed failure reason. Available only if status is `failed`.
    `error`:  structured error message. Available only if status is `failed`.
    """

    user_logs: Logs
    status: ExecutionStatus = ExecutionStatus.PENDING
    failure_type: Optional[FailureType] = None
    error: Optional[Error] = None

    @classmethod
    def from_exception(cls, e: ExecFailureException, user_logs: Logs) -> ExecutionState:
        """Translates a ExecFailureException into a failed ExecutionState we can persist.

        We explicitly require `user_logs`, because the exception does not contain that context, so
        this will force us to explicitly think about when we want to persist logs.
        """
        return cls(
            user_logs=user_logs,
            status=ExecutionStatus.FAILED,
            failure_type=e.failure_type,
            error=Error(
                tip=e.tip,
                context=e.context,
            ),
        )

    def mark_as_failure(self, failure_type: FailureType, tip: str, context: str = "") -> None:
        self.status = ExecutionStatus.FAILED
        self.failure_type = failure_type
        self.error = Error(
            tip=tip,
            context=context,
        )

    def user_fn_redirected(self, failure_tip: str) -> Callable[..., Any]:
        """
        Usage:
        ```
        @exec_state.user_fn_redirected(failure_tip="some message when decorated fn failed")
        def user_fn():
            # run some fn user specified

        user_fn()
        ```
        When decorated with `user_fn_redirected`, the stdout and stderr will be redirected
        to `user_logs`.

        When the decorated fn failed, the `exec_state` will be 'failed' with type 'user'.
        The `error` object will contain the first frame of stack together with the tip provided.
        """

        def wrapper(user_fn: Callable[..., Any]) -> Callable[..., Any]:
            def inner(*args: Any, **kwargs: Any) -> Any:
                stdout_log = io.StringIO()
                stderr_log = io.StringIO()
                try:
                    with redirect_stdout(stdout_log), redirect_stderr(stderr_log):
                        result = user_fn(*args, **kwargs)
                except Exception:
                    # Include the stack trace within the user's code.
                    _set_redirected_logs(stdout_log, stderr_log, self.user_logs)
                    self.mark_as_failure(
                        FailureType.USER_FATAL,
                        failure_tip,
                        context=stack_traceback(
                            offset=1
                        ),  # traceback the first stack frame, which belongs to user
                    )
                    print(f"User failure. Full log: {self.json()}")
                    return None

                # Include the stack trace within the user's code.
                _set_redirected_logs(stdout_log, stderr_log, self.user_logs)
                print(f"User execution succeeded. Full log: {self.json()}")
                return result

            return inner

        return wrapper


def _set_redirected_logs(
    stdout: io.StringIO,
    stderr: io.StringIO,
    logs: Logs,
) -> None:
    """
    If there is any output, set as the values for protected keys STDOUT_KEY and STDERR_KEY.
    """
    stdout.seek(0)
    stderr.seek(0)

    stdout_contents = stdout.read()
    if len(stdout_contents) > 0:
        print(f"StdOut: \n {stdout_contents}")
        logs.stdout = stdout_contents

    stderr_contents = stderr.read()
    if len(stderr_contents) > 0:
        print(f"StdErr: \n {stderr_contents}")
        logs.stderr = stderr_contents
    return


def stack_traceback(offset: int = 0) -> str:
    """
    Captures the stack traceback and returns it as a string. If offset is positive,
    it will extract the traceback starting at OFFSET frames from the top (e.g. most recent frame).
    An offset of 1 means the most recent frame will be excluded.

    This is typically used for user function traceback so that we throw away
    unnecessary stack frames.
    """
    file = io.StringIO()

    tb_type, tb_val, tb = sys.exc_info()
    while offset > 0:
        if tb is None or tb.tb_next is None:
            break
        tb = tb.tb_next
        offset -= 1

    traceback.print_exception(tb_type, tb_val, tb, file=file)

    file.seek(0)
    return file.read()


def exception_traceback(exception: Exception) -> str:
    """
    `exception_traceback` prints the traceback of the entire exception.

    This is typically used for system error so that the full trace is captured.
    """
    return (
        "".join(traceback.format_tb(exception.__traceback__))
        + f"{exception.__class__.__name__}: {str(exception)}"
    )
