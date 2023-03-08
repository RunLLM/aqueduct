import argparse
import os
import subprocess
import sys


def _execute_command(args, cwd=None) -> None:
    with subprocess.Popen(args, stdout=sys.stdout, stderr=sys.stderr, cwd=cwd) as proc:
        proc.communicate()
        if proc.returncode != 0:
            raise Exception("Error executing command: %s" % args)


def _run_tests(
    dir_name: str,
    test_case: str,
    concurrency: int,
    rerun_failed: bool,
    skip_data_setup: bool,
    skip_engine_setup: bool,
) -> None:
    """Either test_case or rerun_failed can be set, but not both."""
    if rerun_failed:
        cmd = ["pytest", dir_name, "-rP", "-vv", "--lf", "-n", str(concurrency)]
    else:
        cmd = ["pytest", dir_name, "-rP", "-vv", "-n", str(concurrency)]

    if len(test_case) > 0:
        cmd += ["-k", test_case]

    if skip_data_setup:
        cmd.append("--skip-data-setup")
    if skip_engine_setup:
        cmd.append("--skip-engine-setup")

    _execute_command(cmd)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()

    parser.add_argument(
        "--data-integration",
        dest="data_integration_tests",
        default=False,
        action="store_true",
        help="Run the SDK Data Integration tests.",
    )

    parser.add_argument(
        "--aqueduct",
        dest="aqueduct_tests",
        default=False,
        action="store_true",
        help="Run the SDK Aqueduct tests.",
    )

    parser.add_argument(
        "-k",
        dest="test_case",
        default="",
        action="store",
        help="Only runs tests that match this argument",
    )

    parser.add_argument(
        "--lf",
        dest="rerun_failed",
        default=False,
        action="store_true",
        help="Run only the tests in the suite that failed during the last run.",
    )

    parser.add_argument(
        "--skip-data-setup",
        dest="skip_data_setup",
        default=False,
        action="store_true",
        help="If set, skips any data integration setup to speed up testing.",
    )

    parser.add_argument(
        "--skip-engine-setup",
        dest="skip_engine_setup",
        default=False,
        action="store_true",
        help="If set, skips any engine integration setup to speed up testing.",
    )

    parser.add_argument(
        "-n",
        dest="concurrency",
        default=8,
        action="store",
        help="The concurrency to run the test suite with.",
    )

    args = parser.parse_args()
    if not (args.aqueduct_tests or args.data_integration_tests):
        args.aqueduct_tests = True
        args.data_integration_tests = True

    assert not (
        args.rerun_failed and len(args.test_case) > 0
    ), "Either -k or -lf can be set, but not both."

    cwd = os.getcwd()
    if not cwd.endswith("integration_tests/sdk"):
        print("Current directory should be the SDK integration test directory.")
        print("Your working directory is %s" % cwd)
        exit(1)

    if args.aqueduct_tests:
        print("Running Aqueduct Tests...")
        _run_tests(
            "aqueduct_tests/",
            args.test_case,
            args.concurrency,
            args.rerun_failed,
            args.skip_data_setup,
            args.skip_engine_setup,
        )

    if args.data_integration_tests:
        print("Running Data Integration Tests...")
        _run_tests(
            "data_integration_tests/",
            args.test_case,
            args.concurrency,
            args.rerun_failed,
            args.skip_data_setup,
            args.skip_engine_setup,
        )
