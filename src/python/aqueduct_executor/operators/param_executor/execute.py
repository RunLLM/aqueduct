import sys

from aqueduct.serialization import deserialize
from aqueduct.utils import infer_artifact_type
from aqueduct_executor.operators.param_executor.spec import ParamSpec
from aqueduct_executor.operators.utils import enums, utils
from aqueduct_executor.operators.utils.execution import (
    TIP_UNKNOWN_ERROR,
    Error,
    ExecFailureException,
    ExecutionState,
    Logs,
    exception_traceback,
)
from aqueduct_executor.operators.utils.storage.parse import parse_storage


def run(spec: ParamSpec) -> None:
    """Parameter operators are unique in that the output content is expected to have already been populated correctly.

    Therefore, this operator is responsible only for:
    - Checking that the parameter type matches the expected type.
    - Writing the operator and artifact metadata to storage, so that orchestration can proceed as normal.

    The artifact output paths are written to before any type checking occurs, so all artifact-related paths are
    expected to have been populated, even if the operator itself fails. However, this is not guaranteed to be the case,
    since system errors are still possible.
    """
    print("Job Spec: \n{}".format(spec.json()))

    storage = parse_storage(spec.storage_config)

    try:
        val = deserialize(
            spec.serialization_type,
            spec.expected_type,
            storage.get(spec.output_content_path),
        )
        inferred_type = infer_artifact_type(val)

        # This does not write to the output artifact's content path as a performance optimization.
        # That has already been written by the Golang Orchestrator.
        utils.write_artifact(
            storage,
            inferred_type,
            None,  # output_content_path
            spec.output_metadata_path,
            val,
            system_metadata={},
        )

        if inferred_type != spec.expected_type:
            raise ExecFailureException(
                failure_type=enums.FailureType.USER_FATAL,
                tip="Supplied parameter expects type `%s`, but got `%s` instead."
                % (spec.expected_type, inferred_type),
            )

        utils.write_exec_state(
            storage,
            spec.metadata_path,
            ExecutionState(status=enums.ExecutionStatus.SUCCEEDED, user_logs=Logs()),
        )

    except ExecFailureException as e:
        from_exception_exec_state = ExecutionState.from_exception(e, user_logs=Logs())
        print(f"Failed with error. Full Logs:\n{from_exception_exec_state.json()}")
        utils.write_exec_state(storage, spec.metadata_path, from_exception_exec_state)
        sys.exit(1)
    except Exception as e:
        exec_state = ExecutionState(
            status=enums.ExecutionStatus.FAILED,
            failure_type=enums.FailureType.SYSTEM,
            error=Error(context=exception_traceback(e), tip=TIP_UNKNOWN_ERROR),
            user_logs=Logs(),
        )
        print(f"Failed with system error. Full Logs:\n{exec_state.json()}")
        utils.write_exec_state(storage, spec.metadata_path, exec_state)
        sys.exit(1)
