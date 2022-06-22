import io
import sys
import traceback

from contextlib import redirect_stderr, redirect_stdout
from typing import Callable, Dict, Optional
from pydantic import BaseModel
from aqueduct_executor.operators.utils.enums import ExecutionCode
from aqueduct_executor.operators.utils import utils
from aqueduct_executor.operators.utils.storage.storage import Storage


_GITHUB_ISSUE_LINK = "https://github.com/aqueducthq/aqueduct/issues/new?assignees=&labels=bug&template=bug_report.md&title=%5BBUG%5D"

TIP_OP_EXECUTION = "Error executing operator. Please refer to the stack trace for fix."
_TIP_CREATE_BUG_REPORT = (
    "We are sorry to see this :(. "
    f"You could send over a bug report through github issue {_GITHUB_ISSUE_LINK} "
    " or in our slack channel. We will get back to you as soon as we can."
)
TIP_UNKNOWN_ERROR = f"An unexpected error occurred. {_TIP_CREATE_BUG_REPORT}"
TIP_INTEGRATION_CONNECTION = (
    "We have trouble connecting to the integration. "
    "Please check your credentials or your integraiton provider."
)
TIP_DEMO_CONNECTION = "We have trouble connecting to demo DB. {_TIP_CREATE_BUG_REPORT}"

TIP_EXTRACT = "We couldn't execute the provided query. Please double check your query is correct."
TIP_LOAD = "We couldn't load to the integration. Please make sure the target exists, or you have the right permission."
TIP_DISCOVER = "We couldn't list items in the integration. Please make sure your credentials have the right permission."

class Error(BaseModel):
    context: str = ""
    tip: str = ""

class Logs(BaseModel):
    stdout: str = ""
    stderr: str = ""
    error: Optional[Error] = None

class Logger(BaseModel):
    all_logs: Dict[str, Logs]
    code: ExecutionCode

def _fetch_redirected_logs(
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

def _user_fn_traceback(offset: int = 0) -> str:
    """
    Captures the stack traceback and returns it as a string. If offset is positive,
    it will extract the traceback starting at OFFSET frames from the top (e.g. most recent frame).
    An offset of 1 means the most recent frame will be excluded.
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

"""
@logged is a decorator which can be used to explictly redirect stdout and stderr
to a `logger` object. When the decorated funciton throws an error, it will also
generate the error message and stack trace.

Args:
logger: the logger object to store stdout, stderr, and error messages.
        Typically, each operator execution pod should have exactly one logger.
is_user_fn: whether the fn is a user fn. This flag controls how stack trace is obtained.
failure_code:

Usage:
@logged(
    logger=logger,
    is_user_fn=True,
)
def f(x, y, z):
    print(x)
    raise Exception("intentional")

f(x, y, z)
"""
def logged(
    logger: Logger,
    key: str,
    failure_code: ExecutionCode,
    failure_tip: str,
    is_user_fn: bool,
    mark_success: bool = False,
    upload_logs_path: str = "",
    storage: Optional[Storage] = None,
):
    def wrapper(fn: Callable) -> Callable:
        def inner(*args, **kwargs):
            stdout_log = io.StringIO()
            stderr_log = io.StringIO()
            result = None
            try:
                with redirect_stdout(stdout_log), redirect_stderr(stderr_log):
                    result = fn(*args, **kwargs)
                logs = logger.user_logs if is_user_fn else logger.system_logs
                _fetch_redirected_logs(stdout_log, stderr_log, logs)
                if mark_success:
                    logger.code = ExecutionCode.SUCCEEDED
                
                if storage and upload_logs_path:
                    utils.write_operator_metadata(storage, upload_logs_path, logger)
                return result
            except Exception as e:
                logs = logger.user_logs if is_user_fn else logger.system_logs
                logger.code = failure_code
                _fetch_redirected_logs(stdout_log, stderr_log, logs)
                ctx = _user_fn_traceback(offset=1) if is_user_fn else ''.join(traceback.format_tb(e.__traceback__))
                logs.error = Error(
                    context=ctx,
                    tip=failure_tip,
                )
                
                if storage and upload_logs_path:
                    utils.write_operator_metadata(storage, upload_logs_path, logger)
                    sys.exit(1)
            return result
        return inner
    return wrapper

