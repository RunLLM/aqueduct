import argparse
import os
import subprocess
import sys
from typing import List, Optional


def _execute_command(args, cwd=None) -> None:
    with subprocess.Popen(args, stdout=sys.stdout, stderr=sys.stderr, cwd=cwd) as proc:
        proc.communicate()
        if proc.returncode != 0:
            raise Exception("Error executing command: %s" % args)


def _run_tests(
    dir_name: str,
    file_name: Optional[str],
    concurrency: int,
    unknown_args: List[str],
) -> None:
    """Either test_case or rerun_failed can be set, but not both."""
    target_name = dir_name
    if file_name is not None:
        # `dir_name` already ends in a slash.
        assert dir_name[-1] == "/"
        target_name = dir_name + file_name

    cmd = ["pytest", target_name, "-rP", "-vv"] + unknown_args + ["-n", str(concurrency)]
    _execute_command(cmd)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()

    parser.add_argument(
        "--data",
        dest="data_resource_tests",
        default=False,
        action="store_true",
        help="Run the SDK Data Resource tests.",
    )

    parser.add_argument(
        "--aqueduct",
        dest="aqueduct_tests",
        default=False,
        action="store_true",
        help="Run the SDK Aqueduct tests.",
    )

    parser.add_argument(
        "--file",
        dest="file",
        default=None,
        action="store",
        help="The file to run the tests on. For example, `python3 run_tests.py --aqueduct --file flow_test.py` is "
        "equivalent to running `pytest aqueduct/flow_test.py`.",
    )

    parser.add_argument(
        "-n",
        dest="concurrency",
        default=8,
        action="store",
        help="The concurrency to run the test suite with.",
    )

    args, unknown_args = parser.parse_known_args()
    if not (args.aqueduct_tests or args.data_resource_tests):
        args.aqueduct_tests = True
        args.data_resource_tests = True

    cwd = os.getcwd()
    if not cwd.endswith("integration_tests/sdk"):
        print("Current directory should be the SDK integration test directory.")
        print("Your working directory is %s" % cwd)
        exit(1)

    if args.aqueduct_tests:
        print("Running Aqueduct Tests...")
        _run_tests(
            "aqueduct_tests/",
            args.file,
            args.concurrency,
            unknown_args,
        )

    if args.data_resource_tests:
        print("Running Data Resource Tests...")
        _run_tests(
            "data_resource_tests/",
            args.file,
            args.concurrency,
            unknown_args,
        )
