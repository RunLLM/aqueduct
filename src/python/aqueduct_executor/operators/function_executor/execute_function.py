import argparse
import base64
import importlib
import io
import json
import os
import sys
import traceback
from contextlib import redirect_stderr, redirect_stdout
from typing import Any, Callable, List

from aqueduct_executor.operators.function_executor import spec
from aqueduct_executor.operators.function_executor.utils import OP_DIR
from aqueduct_executor.operators.utils import utils
from aqueduct_executor.operators.utils.enums import ExecutionCode
from aqueduct_executor.operators.utils.logging import ExecutionStatus, Logs, Error, TIP_OP_EXECUTION, TIP_UNKNOWN_ERROR
from aqueduct_executor.operators.utils.storage.parse import parse_storage

from pandas import DataFrame


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


def _get_py_import_path(spec: spec.FunctionSpec) -> str:
    """
    Generates the import path based on fixed function dir and
    FUNCTION_ENTRY_POINT_FILE env var.

    It removes .py (if any) from the entry point and replaces all
    '/' with '.'

    For example, entry point 'model/churn.py'  will finally become
    'app.function.model.churn', where we can import from.
    """
    file_path = spec.entry_point_file
    if file_path.endswith(".py"):
        file_path = file_path[:-3]

    if file_path.startswith("/"):
        file_path = file_path[1:]
    return ".".join([OP_DIR, file_path.replace("/", ".")])


def _import_invoke_method(spec: spec.FunctionSpec) -> Callable[..., DataFrame]:
    fn_path = spec.function_extract_path
    os.chdir(os.path.join(fn_path, OP_DIR))
    sys.path.append(fn_path)
    import_path = _get_py_import_path(spec)
    class_name = spec.entry_point_class
    method_name = spec.entry_point_method
    custom_args_str = spec.custom_args
    # Invoke the function and parse out the result object.
    module = importlib.import_module(import_path)
    if not class_name:
        return getattr(module, method_name)

    fn_class = getattr(module, class_name)
    function = fn_class()
    # Set the custom arguments if provided
    if custom_args_str:
        custom_args = json.loads(custom_args_str)
        function.set_args(custom_args)

    return getattr(function, method_name)


def _execute_function(
    spec: spec.FunctionSpec,
    inputs: List[utils.InputArtifact],
    exec_status: ExecutionStatus,
) -> Any:
    """
    Invokes the given function on the input data. Does not raise an exception on any
    user function errors. Instead, returns the error message as a string.

    :param inputs: the input data to feed into the user's function.
    """
    stdout_log = io.StringIO()
    stderr_log = io.StringIO()

    invoke = _import_invoke_method(spec)
    print("Invoking the function...")
    result = None
    try:
        with redirect_stdout(stdout_log), redirect_stderr(stderr_log):
            result = invoke(*inputs)  # Unpack DataFrames argument list
    except Exception:
        # Include the stack trace within the user's code.
        sys.path.pop(0)
        exec_status.code = ExecutionCode.USER_FAILURE
        _fetch_redirected_logs(stdout_log, stderr_log, exec_status.user_logs)
        exec_status.user_logs.error = Error(
            context=_user_fn_traceback(offset=1),
            tip=TIP_OP_EXECUTION,
        )
        return None

    sys.path.pop(0)
    return result


def run(spec: spec.FunctionSpec) -> None:
    """
    Executes a function operator.
    """
    
    exec_status = ExecutionStatus(
        user_logs=Logs(),
        system_logs=Logs(),
        code=ExecutionCode.UNKNOWN,
    )
    storage = parse_storage(spec.storage_config)
    stdout_log = io.StringIO()
    stderr_log = io.StringIO()
    try:
        # Read the input data from intermediate storage.
        with redirect_stdout(stdout_log), redirect_stderr(stderr_log):
            inputs = utils.read_artifacts(
                storage, spec.input_content_paths, spec.input_metadata_paths, spec.input_artifact_types
            )

            print("Invoking the function...")
            results = _execute_function(spec, inputs, exec_status)
            if exec_status.code == ExecutionCode.USER_FAILURE:
                _fetch_redirected_logs(stdout_log, stderr_log, exec_status.system_logs)
                utils.write_execution_status(storage, spec.metadata_path, exec_status)
                sys.exit(1)

            print("Function invoked successfully!")

            # Force all results to be of type `list`, so we can always loop over them.
            if not isinstance(results, list):
                results = [results]

            utils.write_artifacts(
                storage,
                spec.output_content_paths,
                spec.output_metadata_paths,
                results,
                spec.output_artifact_types,
            )
        
        _fetch_redirected_logs(stdout_log, stderr_log, exec_status.system_logs)
        exec_status.code = ExecutionCode.SUCCEEDED 
        utils.write_operator_metadata(storage, spec.metadata_path, exec_status)

    except Exception as e:
        _fetch_redirected_logs(stdout_log, stderr_log, exec_status.system_logs)
        exec_status.code = ExecutionCode.SYSTEM_FAILURE
        exec_status.system_logs.error = Error(
            context=''.join(traceback.format_tb(e.__traceback__)),
            tip=TIP_UNKNOWN_ERROR,
        )
        utils.write_operator_metadata(storage, spec.metadata_path, exec_status)
        sys.exit(1)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-s", "--spec", required=True)
    args = parser.parse_args()

    spec_json = base64.b64decode(args.spec)
    spec = spec.parse_spec(spec_json)

    print("Started %s job: %s" % (spec.type, spec.name))

    run(spec)
