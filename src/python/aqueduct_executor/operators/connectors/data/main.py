import argparse
import base64
import sys

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-s", "--spec", required=True)
    parser.add_argument("-v", "--version-tag", default="")
    args = parser.parse_args()
    if args.version_tag:
        import subprocess

        install_process = subprocess.run(
            [
                sys.executable,
                "-m",
                "pip",
                "install",
                "--index-url",
                "https://test.pypi.org/simple/",
                "--extra-index-url",  # allows dependencies from pypi
                "https://pypi.org/simple",
                f"aqueduct-ml=={args.version_tag}",
            ]
        )
        print(install_process.stderr)
        print(install_process.stdout)
        install_process.check_returncode()

    from aqueduct_executor.operators.connectors.data import execute
    from aqueduct_executor.operators.connectors.data.spec import parse_spec
    from aqueduct_executor.operators.utils.utils import time_it

    spec_json = base64.b64decode(args.spec)
    spec = parse_spec(spec_json)

    time_it(job_name=spec.name, job_type=spec.type.value, step="Running Connector")(execute.run)(
        spec
    )
