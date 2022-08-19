import importlib
import json
import os
import shutil
import sys
import tracemalloc
import uuid
from typing import Any, Callable, Dict, List, Tuple

import cloudpickle as pickle
import numpy as np
import pandas as pd
from aqueduct_executor.operators.function_executor import extract_function, get_extract_path
from aqueduct_executor.operators.function_executor.spec import FunctionSpec
from aqueduct_executor.operators.function_executor.utils import OP_DIR
from aqueduct_executor.operators.utils import utils
from aqueduct_executor.operators.utils.enums import (
    ArtifactType,
    CheckSeverityLevel,
    ExecutionStatus,
    FailureType,
    OperatorType,
)
from aqueduct_executor.operators.utils.execution import (
    TIP_CHECK_DID_NOT_PASS,
    TIP_NOT_BOOL,
    TIP_NOT_NUMERIC,
    TIP_OP_EXECUTION,
    TIP_UNKNOWN_ERROR,
    Error,
    ExecutionState,
    Logs,
    exception_traceback,
)
from aqueduct_executor.operators.utils.storage.parse import parse_storage
from aqueduct_executor.operators.utils.timer import Timer
from aqueduct_executor.operators.utils.utils import check_passed, infer_artifact_type
from pandas import DataFrame
from PIL import Image


def _get_py_import_path(spec: FunctionSpec) -> str:
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
    return file_path.replace("/", ".")


def _import_invoke_method(spec: FunctionSpec) -> Callable[..., Any]:
    """
    `_import_invoke_method` imports the model object.
    it assumes the operator has been extracted to `<storage>/operators/<id>/op`
    and imports the route from the above path.
    """

    # fn_path should be `<storage>/operators/<id>`
    fn_path = spec.function_extract_path

    # work_dir should be `<storage>/operators/<id>/op`
    work_dir = os.path.join(fn_path, OP_DIR)
    print(f"listdir(workdir): {os.listdir(work_dir)}")
    print(f"listdir(fn_path): {os.listdir(fn_path)}")

    # this ensures any file manipulation happens with respect to work_dir
    os.chdir(work_dir)
    # adds work_dir to sys.path to support relative imports from work_dir
    sys.path.append(work_dir)

    import_path = _get_py_import_path(spec)
    print(f"import_path: {import_path}")
    class_name = spec.entry_point_class
    method_name = spec.entry_point_method
    custom_args_str = spec.custom_args

    # Invoke the function and parse out the result object.

    module = importlib.import_module(import_path)

    if not class_name:
        return getattr(module, method_name)  # type: ignore

    fn_class = getattr(module, class_name)
    function = fn_class()
    # Set the custom arguments if provided
    if custom_args_str:
        custom_args = json.loads(custom_args_str)
        function.set_args(custom_args)

    return getattr(function, method_name)  # type: ignore


def _execute_function(
    spec: FunctionSpec,
    inputs: List[Any],
    exec_state: ExecutionState,
) -> Tuple[Any, ArtifactType, Dict[str, str]]:
    """
    Invokes the given function on the input data. Does not raise an exception on any
    user function errors. Instead, returns the error message as a string.

    :param inputs: the input data to feed into the user's function.
    """

    invoke = _import_invoke_method(spec)
    timer = Timer()
    print("Invoking the function...")
    timer.start()
    tracemalloc.start()

    @exec_state.user_fn_redirected(failure_tip=TIP_OP_EXECUTION)
    def _invoke() -> Any:
        return invoke(*inputs)

    result = _invoke()
    inferred_result_type = infer_artifact_type(result)

    elapsedTime = timer.stop()
    _, peak = tracemalloc.get_traced_memory()
    system_metadata = {
        utils._RUNTIME_SEC_METRIC_NAME: str(elapsedTime),
        utils._MAX_MEMORY_MB_METRIC_NAME: str(peak / 10**6),
    }

    sys.path.pop(0)
    return result, inferred_result_type, system_metadata


def run(spec: FunctionSpec) -> None:
    """
    Executes a function operator.
    """
    print("Started %s job: %s" % (spec.type, spec.name))

    exec_state = ExecutionState(user_logs=Logs())
    storage = parse_storage(spec.storage_config)
    try:
        # Read the input data from intermediate storage.
        inputs, _ = utils.read_artifacts(
            storage, spec.input_content_paths, spec.input_metadata_paths
        )

        print("Invoking the function...")
        result, result_type, system_metadata = _execute_function(spec, inputs, exec_state)
        if exec_state.status == ExecutionStatus.FAILED:
            # user failure
            utils.write_exec_state(storage, spec.metadata_path, exec_state)
            sys.exit(1)

        print("Function invoked successfully!")

        # Perform type checking for metric and check operators.
        type_error = False
        if spec.operator_type == OperatorType.METRIC:
            if not (
                isinstance(result, int)
                or isinstance(result, float)
                or isinstance(result, np.number)
            ):
                type_error = True
                type_error_tip = TIP_NOT_NUMERIC
        elif spec.operator_type == OperatorType.CHECK:
            if isinstance(result, pd.Series) and result.dtype == "bool":
                # Cast pd.Series to a bool.
                # We only write True if every boolean in the series is True.
                series = pd.Series(result)
                result = bool(series.size - series.sum().item() == 0)
                result_type = ArtifactType.BOOL
            elif isinstance(result, bool) or isinstance(result, np.bool_):
                # Cast np.bool_ to a bool.
                result = bool(result)
            else:
                type_error = True
                type_error_tip = TIP_NOT_BOOL

        if type_error:
            exec_state.status = ExecutionStatus.FAILED
            exec_state.failure_type = FailureType.USER_FATAL
            exec_state.error = Error(
                context="",
                tip=type_error_tip,
            )
            utils.write_exec_state(storage, spec.metadata_path, exec_state)
            sys.exit(1)

        utils.write_artifact(
            storage,
            result_type,
            spec.output_content_paths[0],
            spec.output_metadata_paths[0],
            result,
            system_metadata=system_metadata,
        )

        # For check operators, we want to fail the operator based on the exact output of the user's function.
        # Assumption: the check operator only has a single output.
        if spec.operator_type == OperatorType.CHECK and not check_passed(result):
            check_severity = spec.check_severity
            if spec.check_severity is None:
                print("Check operator has an unspecified severity on spec. Defaulting to ERROR.")
                check_severity = CheckSeverityLevel.ERROR

            failure_type = FailureType.USER_FATAL
            if check_severity == CheckSeverityLevel.WARNING:
                failure_type = FailureType.USER_NON_FATAL

            exec_state.status = ExecutionStatus.FAILED
            exec_state.failure_type = failure_type
            exec_state.error = Error(
                context="",
                tip=TIP_CHECK_DID_NOT_PASS,
            )
            utils.write_exec_state(storage, spec.metadata_path, exec_state)
            print(f"Check Operator did not pass. Full logs: {exec_state.json()}")
        else:
            exec_state.status = ExecutionStatus.SUCCEEDED
            utils.write_exec_state(storage, spec.metadata_path, exec_state)
            print(f"Succeeded! Full logs: {exec_state.json()}")

    except Exception as e:
        exec_state.status = ExecutionStatus.FAILED
        exec_state.failure_type = FailureType.SYSTEM
        exec_state.error = Error(
            context=exception_traceback(e),
            tip=TIP_UNKNOWN_ERROR,
        )
        print(f"Failed with system error. Full Logs:\n{exec_state.json()}")
        utils.write_exec_state(storage, spec.metadata_path, exec_state)
        sys.exit(1)


def run_with_setup(spec: FunctionSpec) -> None:
    """
    Performs the setup needed for a Function operator and then executes it.
    """
    # Generate a unique function extract path if one does not exist already
    fn_extract_path = None
    if not spec.function_extract_path:
        fn_extract_path = os.path.join(os.getcwd(), str(uuid.uuid4()))
        spec.function_extract_path = fn_extract_path

    op_path = get_extract_path.run(spec)

    extract_function.run(spec)

    requirements_path = os.path.join(op_path, "requirements.txt")
    if os.path.exists(requirements_path):
        os.system("{} -m pip install -r {}".format(sys.executable, requirements_path))

    run(spec)

    if fn_extract_path:
        # Delete extracted function
        shutil.rmtree(fn_extract_path)
