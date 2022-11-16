import argparse
import base64
import subprocess
import sys

from aqueduct_executor.operators.function_executor.spec import FunctionSpec, parse_spec
from aqueduct_executor.operators.utils import utils
from aqueduct_executor.operators.utils.enums import FailureType
from aqueduct_executor.operators.utils.execution import ExecFailureException, ExecutionState, Logs
from aqueduct_executor.operators.utils.storage.parse import parse_storage


def install_missing_packages(missing_path: str, spec: FunctionSpec) -> None:
    install_output = subprocess.run(
        [sys.executable, "-m", "pip", "install", "-r", missing_path], capture_output=True, text=True
    )

    if install_output.returncode != 0:
        exception = ExecFailureException(
            failure_type=FailureType.USER_FATAL,
            tip="We are unable to install certain dependency packages. Please remove them from the \
requirement file and try again. Please refer to the stderr log for which package \
caused the installation error.",
        )
        from_exception_exec_state = ExecutionState.from_exception(
            exception, user_logs=Logs(stdout=install_output.stdout, stderr=install_output.stderr)
        )

        utils.write_exec_state(
            parse_storage(spec.storage_config), spec.metadata_path, from_exception_exec_state
        )
        sys.exit(1)


def run(local_path: str, requirements_path: str, missing_path: str, spec: FunctionSpec) -> None:
    with open(local_path, "r") as f:
        local_req = set(f.read().split("\n"))

    with open(requirements_path, "r") as f:
        required = f.read().split("\n")

    missing = []
    for r in required:
        # Remove any @ file because we may not have those files local to the user's device in our file system.
        if r not in local_req and "@ file" not in r:
            missing.append(r)

    if len(missing) > 0:
        with open(missing_path, "w") as f:
            f.write("\n".join(missing))
        install_missing_packages(missing_path, spec)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--local_path", required=True)
    parser.add_argument("--requirements_path", required=True)
    parser.add_argument("--missing_path", required=True)
    parser.add_argument("--spec", required=True)
    args = parser.parse_args()

    spec_json = base64.b64decode(args.spec)
    spec = parse_spec(spec_json)

    run(args.local_path, args.requirements_path, args.missing_path, spec)
