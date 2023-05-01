import importlib
import json
import os
import shutil
import subprocess
import sys
import tracemalloc
import uuid
from typing import Any, Callable, Dict, List, Tuple

import numpy as np
import pandas as pd
from aqueduct.utils.serialization import check_and_fetch_pickled_collection_format
from aqueduct.utils.type_inference import infer_artifact_type
from aqueduct_executor.operators.function_executor import extract_function, get_extract_path
from aqueduct_executor.operators.function_executor.spec import FunctionSpec
from aqueduct_executor.operators.function_executor.utils import OP_DIR
from aqueduct_executor.operators.utils import utils
from aqueduct_executor.operators.utils.enums import (
    ArtifactType,
    CheckSeverity,
    ExecutionStatus,
    FailureType,
    OperatorType,
    SerializationType,
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
from aqueduct_executor.operators.utils.storage.storage import Storage
from aqueduct_executor.operators.utils.timer import Timer
from aqueduct_executor.operators.utils.utils import time_it


def get_py_import_path(spec: FunctionSpec) -> str:
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


def import_invoke_method(spec: FunctionSpec) -> Callable[..., Any]:
    """
    `import_invoke_method` imports the model object.
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

    import_path = get_py_import_path(spec)
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
) -> Tuple[List[Any], Dict[str, str]]:
    """
    Invokes the given function on the input data. Does not raise an exception on any
    user function errors, but instead annotates the given exec state with the error.

    :param inputs: the input data to feed into the user's function.
    """

    invoke = import_invoke_method(spec)
    timer = Timer()
    timer.start()
    tracemalloc.start()

    @exec_state.user_fn_redirected(failure_tip=TIP_OP_EXECUTION)
    def _invoke() -> Any:
        return invoke(*inputs)

    results = _invoke()
    if len(spec.output_content_paths) == 1:
        results = [results]

    elapsedTime = timer.stop()
    _, peak = tracemalloc.get_traced_memory()
    system_metadata = {
        utils._RUNTIME_SEC_METRIC_NAME: str(elapsedTime),
        utils._MAX_MEMORY_MB_METRIC_NAME: str(peak / 10**6),
    }

    sys.path.pop(0)
    return results, system_metadata


def _validate_result_count_and_infer_type(
    spec: FunctionSpec,
    results: List[Any],
    infer_type_func: Any,
) -> List[ArtifactType]:
    """
    Validates that the expected number of results were returned by the Function
    and infers the ArtifactType of each result.

    Args:
        spec: The FunctionSpec for the Function
        results: The results returned by the Function

    Returns:
        The ArtifactType of each result

    Raises:
        ExecFailureException: If the expected number of results were not returned
    """
    if len(spec.output_content_paths) > 1 and len(spec.output_content_paths) != len(results):
        raise ExecFailureException(
            failure_type=FailureType.USER_FATAL,
            tip="Expected function to have %s outputs, but instead it had %s."
            % (len(spec.output_content_paths), len(results)),
        )

    return [infer_type_func(res) for res in results]


def _write_artifacts(
    write_artifact_func: Any,
    results: Any,
    result_types: List[ArtifactType],
    derived_from_bson: bool,
    output_content_paths: List[str],
    output_metadata_paths: List[str],
    system_metadata: Any,
    storage: Storage,
    **kwargs: Any,
) -> None:
    for i, result in enumerate(results):
        write_artifact_func(
            storage,
            result_types[i],
            derived_from_bson,
            output_content_paths[i],
            output_metadata_paths[i],
            result,
            system_metadata=system_metadata,
            **kwargs,
        )


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


def cleanup(spec: FunctionSpec) -> None:
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
    execute_function_spec(
        spec=spec,
        read_artifacts_func=utils.read_artifacts,
        write_artifact_func=utils.write_artifact,
        infer_type_func=infer_artifact_type,
    )


def execute_function_spec(
    spec: FunctionSpec,
    read_artifacts_func: Any,
    write_artifact_func: Any,
    infer_type_func: Any,
    **kwargs: Any,
) -> None:
    """
    Executes a function operator. If run in a Spark environment, it uses the Spark specific utils
    functions to read/write to storage layer and to infer the type of artifact.
    The only kwarg we expect is spark_session_obj.

    Args:
        spec: The spec provided for this operator.
        read_artifacts_func: function used to read artifacts from storage layer
        write_artifact_func: function used to write artifacts to storage layer
        infer_type_func: function used to infer type of artifacts returned by operators.
    """
    exec_state = ExecutionState(user_logs=Logs())
    storage = parse_storage(spec.storage_config)
    try:
        check_package_version_mismatch()

        validate_spec(spec)

        # Read the input data from intermediate storage.
        inputs, _, serialization_types = time_it(
            job_name=spec.name, job_type=spec.type.value, step="Reading Inputs"
        )(read_artifacts_func)(
            storage=storage,
            input_paths=spec.input_content_paths,
            input_metadata_paths=spec.input_metadata_paths,
            **kwargs,
        )

        # We need to check for BSON_TABLE serialization type at both the top level
        # and within any serialized pickled collection (if it exists).
        derived_from_bson = SerializationType.BSON_TABLE in serialization_types
        if not derived_from_bson:
            for i, serialization_type in enumerate(serialization_types):
                collection_data = check_and_fetch_pickled_collection_format(
                    serialization_type, inputs[i]
                )
                if (
                    collection_data is not None
                    and SerializationType.BSON_TABLE in collection_data.aqueduct_serialization_types
                ):
                    derived_from_bson = True
                    break

        results, system_metadata = time_it(
            job_name=spec.name, job_type=spec.type.value, step="Running Function"
        )(_execute_function)(spec, inputs, exec_state)

        if exec_state.status == ExecutionStatus.FAILED:
            # user failure
            utils.write_exec_state(storage, spec.metadata_path, exec_state)
            sys.exit(1)

        print("Function invoked successfully!")

        result_types = _validate_result_count_and_infer_type(
            spec=spec, results=results, infer_type_func=infer_type_func
        )

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
                write_artifact_func(
                    storage=storage,
                    artifact_type=ArtifactType.BOOL,
                    derived_from_bson=derived_from_bson,  # derived_from_bson doesn't apply to bool artifact
                    output_path=spec.output_content_paths[0],
                    output_metadata_path=spec.output_metadata_paths[0],
                    content=check_passed,
                    system_metadata=system_metadata,
                    **kwargs,
                )

                check_severity = spec.check_severity
                if spec.check_severity is None:
                    print(
                        "Check operator has an unspecified severity on spec. Defaulting to ERROR."
                    )
                    check_severity = CheckSeverity.ERROR

                failure_type = FailureType.USER_FATAL
                if check_severity == CheckSeverity.WARNING:
                    failure_type = FailureType.USER_NON_FATAL

                raise ExecFailureException(failure_type, tip=TIP_CHECK_DID_NOT_PASS)

            # If we get here, we know that the check has passed. The artifact type might need
            # still be updated. Eg. if the output was a pandas series.
            result_types[0] = ArtifactType.BOOL
            results[0] = True
        else:
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

        time_it(job_name=spec.name, job_type=spec.type.value, step="Writing Outputs")(
            _write_artifacts
        )(
            write_artifact_func=write_artifact_func,
            results=results,
            result_types=result_types,
            derived_from_bson=derived_from_bson,
            output_content_paths=spec.output_content_paths,
            output_metadata_paths=spec.output_metadata_paths,
            system_metadata=system_metadata,
            storage=storage,
            **kwargs,
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
        cleanup(spec)


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


def check_package_version_mismatch() -> None:
    expected_version = os.environ.get("AQUEDUCT_EXPECTED_VERSION")
    if expected_version:
        aqueduct_version = subprocess.check_output(["aqueduct", "version"]).decode("utf-8").strip()

        print(
            f"Comparing Aqueduct version ({aqueduct_version}) to expected version ({expected_version})"
        )
        if aqueduct_version != expected_version:
            raise ExecFailureException(
                failure_type=FailureType.USER_FATAL,
                tip=f"Aqueduct version ({aqueduct_version}) does not match expected version ({expected_version})",
            )
