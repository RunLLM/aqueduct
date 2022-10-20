import importlib
import json
import os
import shutil
import sys
import tracemalloc
import uuid
from typing import Any, Callable, Dict, List, Tuple

import numpy as np
import pandas as pd
from aqueduct.utils import infer_artifact_type
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
    ExecFailureException,
    ExecutionState,
    Logs,
    exception_traceback,
)
from aqueduct_executor.operators.utils.storage.parse import parse_storage
from aqueduct_executor.operators.utils.timer import Timer


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
) -> Tuple[List[Any], List[ArtifactType], Dict[str, str]]:
    """
    Invokes the given function on the input data. Does not raise an exception on any
    user function errors, but instead annotates the given exec state with the error.

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

    results = _invoke()

    # We only validate the number of results if multiple outputs are expected.
    # Otherwise, we treat the results as a single object.
    if len(spec.output_content_paths) > 1 and len(spec.output_content_paths) != len(results):
        raise ExecFailureException(
            failure_type=FailureType.USER_FATAL,
            tip="Expected function to have %s outputs, but instead it had %s."
            % (len(spec.output_content_paths), len(results)),
        )

    if len(spec.output_content_paths) == 1:
        results = [results]

    inferred_result_types = [infer_artifact_type(res) for res in results]

    elapsedTime = timer.stop()
    _, peak = tracemalloc.get_traced_memory()
    system_metadata = {
        utils._RUNTIME_SEC_METRIC_NAME: str(elapsedTime),
        utils._MAX_MEMORY_MB_METRIC_NAME: str(peak / 10**6),
    }

    sys.path.pop(0)
    return results, inferred_result_types, system_metadata


def validate_spec(spec: FunctionSpec) -> None:
    if len(spec.input_content_paths) != len(spec.input_metadata_paths):
        raise Exception(
            "Found inconsistent number of input paths (%d) and input metadata paths (%d)"
            % (
                len(spec.input_content_paths),
                len(spec.input_metadata_paths),
            )
        )

    if len(spec.output_content_paths) != len(spec.output_metadata_paths):
        raise Exception(
            "Found inconsistent number of output paths (%d) and output metadata paths (%d)"
            % (
                len(spec.output_content_paths),
                len(spec.output_metadata_paths),
            )
        )
    if spec.expected_output_artifact_types is not None and len(
        spec.expected_output_artifact_types
    ) != len(spec.output_content_paths):
        raise Exception(
            "Found inconsistent number of expected output artifact types (%d) and output content paths (%d)"
            % (
                len(spec.expected_output_artifact_types),
                len(spec.output_content_paths),
            )
        )

    # Check and Metric operators must only have a single output.
    if (spec.operator_type == OperatorType.CHECK or spec.operator_type == OperatorType.METRIC) and (
        spec.expected_output_artifact_types is not None
        and len(spec.expected_output_artifact_types) != 1
    ):
        raise Exception("%s operators must only have a single output." % spec.operator_type)


def _cleanup(spec: FunctionSpec) -> None:
    """
    Cleans up any temporary files created during function execution.
    """
    # Delete the extracted fn file if it exists and the file path is not
    # something dangerous
    if spec.function_extract_path and spec.function_extract_path[-1] != "*":
        shutil.rmtree(spec.function_extract_path)


def run(spec: FunctionSpec) -> None:
    """
    Executes a function operator.
    """
    print("Started %s job: %s" % (spec.type, spec.name))

    exec_state = ExecutionState(user_logs=Logs())
    storage = parse_storage(spec.storage_config)
    try:
        validate_spec(spec)

        # Read the input data from intermediate storage.
        inputs, _ = utils.read_artifacts(
            storage, spec.input_content_paths, spec.input_metadata_paths
        )

        print("Invoking the function...")
        results, result_types, system_metadata = _execute_function(spec, inputs, exec_state)
        if exec_state.status == ExecutionStatus.FAILED:
            # user failure
            utils.write_exec_state(storage, spec.metadata_path, exec_state)
            sys.exit(1)

        print("Function invoked successfully!")

        # Perform type checking on the function output.
        if spec.operator_type == OperatorType.METRIC:
            assert len(results) == 1, "Metric operator can only have a single output."
            result = results[0]

            if not (
                isinstance(result, int)
                or isinstance(result, float)
                or isinstance(result, np.number)
            ):
                raise ExecFailureException(
                    failure_type=FailureType.USER_FATAL,
                    tip=TIP_NOT_NUMERIC,
                )

        elif spec.operator_type == OperatorType.CHECK:
            assert len(results) == 1, "Check operator can only have a single output."
            check_result = results[0]

            if isinstance(check_result, pd.Series) and check_result.dtype == "bool":
                assert result_types[0] == ArtifactType.PICKLABLE

                # Cast pd.Series to a bool.
                # We only write True if every boolean in the series is True.
                series = pd.Series(check_result)
                check_passed = bool(series.size - series.sum().item() == 0)
            elif isinstance(check_result, bool) or isinstance(check_result, np.bool_):
                # Cast np.bool_ to a bool.
                check_passed = bool(check_result)
            else:
                raise ExecFailureException(
                    failure_type=FailureType.USER_FATAL,
                    tip=TIP_NOT_BOOL,
                )

            # If the check returned a value we interpret to mean 'false', we exit here, but
            # not before recording the output artifact value (which will be False).
            if not check_passed:
                print(f"Check Operator did not pass.")

                utils.write_artifact(
                    storage,
                    ArtifactType.BOOL,
                    spec.output_content_paths[0],
                    spec.output_metadata_paths[0],
                    check_passed,
                    system_metadata=system_metadata,
                )

                check_severity = spec.check_severity
                if spec.check_severity is None:
                    print(
                        "Check operator has an unspecified severity on spec. Defaulting to ERROR."
                    )
                    check_severity = CheckSeverityLevel.ERROR

                failure_type = FailureType.USER_FATAL
                if check_severity == CheckSeverityLevel.WARNING:
                    failure_type = FailureType.USER_NON_FATAL

                raise ExecFailureException(failure_type, tip=TIP_CHECK_DID_NOT_PASS)

            # If we get here, we know that the check has passed. The artifact type might need
            # still be updated. Eg. if the output was a pandas series.
            result_types[0] = ArtifactType.BOOL
            results[0] = True
        else:
            # Error if the number of expected outputs is not what was returned by the user's function.
            if len(results) != len(spec.expected_output_artifact_types):
                raise ExecFailureException(
                    failure_type=FailureType.USER_FATAL,
                    tip="Expected function to return %d outputs, but instead got %d."
                    % (len(spec.expected_output_artifact_types), len(results)),
                )

            for i, expected_output_type in enumerate(spec.expected_output_artifact_types):
                if (
                    expected_output_type != ArtifactType.UNTYPED
                    and expected_output_type != result_types[i]
                ):
                    raise ExecFailureException(
                        failure_type=FailureType.USER_FATAL,
                        tip="Expected type %s for the %d-th output of function, but it is of type %s."
                        % (expected_output_type, i, result_types[i]),
                    )

        for i, result in enumerate(results):
            utils.write_artifact(
                storage,
                result_types[i],
                spec.output_content_paths[i],
                spec.output_metadata_paths[i],
                result,
                system_metadata=system_metadata,
            )

        # If we made it here, then the operator has succeeded.
        exec_state.status = ExecutionStatus.SUCCEEDED
        print(f"Succeeded! Full logs: {exec_state.json()}")
        utils.write_exec_state(storage, spec.metadata_path, exec_state)

    except ExecFailureException as e:
        # We must reconcile the user logs here, since those logs are not captured on the exception.
        from_exception_exec_state = ExecutionState.from_exception(e, user_logs=exec_state.user_logs)
        print(f"Failed with error. Full Logs:\n{from_exception_exec_state.json()}")
        utils.write_exec_state(storage, spec.metadata_path, from_exception_exec_state)
        sys.exit(1)

    except Exception as e:
        exec_state.mark_as_failure(
            FailureType.SYSTEM, TIP_UNKNOWN_ERROR, context=exception_traceback(e)
        )
        print(f"Failed with system error. Full Logs:\n{exec_state.json()}")
        utils.write_exec_state(storage, spec.metadata_path, exec_state)
        sys.exit(1)
    finally:
        # Perform any cleanup
        _cleanup(spec)


def run_with_setup(spec: FunctionSpec) -> None:
    """
    Performs the setup needed for a Function operator and then executes it.
    """
    # Generate a unique function extract path if one does not exist already
    if not spec.function_extract_path:
        fn_extract_path = os.path.join(os.getcwd(), str(uuid.uuid4()))
        spec.function_extract_path = fn_extract_path

    op_path = get_extract_path.run(spec)

    extract_function.run(spec)

    requirements_path = os.path.join(op_path, "requirements.txt")
    if os.path.exists(requirements_path):
        os.system("{} -m pip install -r {}".format(sys.executable, requirements_path))

    run(spec)
