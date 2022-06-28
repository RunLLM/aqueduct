import argparse
import base64
import sys

from aqueduct_executor.operators.utils import enums, utils
from aqueduct_executor.operators.utils.logging import (
    Error,
    ExecutionLogs,
    Logs,
    TIP_UNKNOWN_ERROR,
    exception_traceback,
)
from aqueduct_executor.operators.system_metric_executor import spec
from aqueduct_executor.operators.utils.storage.parse import parse_storage


def run(spec: spec.SystemMetricSpec) -> None:
    """
    Executes a system metric operator by storing the requested system metrics value in the output content path.
    """
    storage = parse_storage(spec.storage_config)
    exec_logs = ExecutionLogs(user_logs=Logs())
    try:
        system_metadata = utils.read_system_metadata(storage, spec.input_metadata_paths)
        utils.write_artifact(
            storage,
            spec.output_content_path,
            spec.output_metadata_path,
            float(system_metadata[0][utils._METADATA_SYSTEM_METADATA_NAME][spec.metric_name]),
            {},
            enums.OutputArtifactType.FLOAT,
        )
        exec_logs.code = enums.ExecutionCode.SUCCEEDED
        utils.write_logs(storage, spec.metadata_path, exec_logs)
    except Exception as e:
        exec_logs.code = enums.ExecutionCode.FAILED
        exec_logs.failure_reason = enums.FailureReason.SYSTEM
        exec_logs.error = Error(context=exception_traceback(e), tip=TIP_UNKNOWN_ERROR)
        print(f"Failed with system error. Full Logs:\n{exec_logs.json()}")
        utils.write_logs(storage, spec.metadata_path, exec_logs)
        sys.exit(1)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-s", "--spec", required=True)
    args = parser.parse_args()

    spec_json = base64.b64decode(args.spec)
    spec = spec.parse_spec(spec_json)

    print("Job Spec: \n{}".format(spec.json()))
    run(spec)
