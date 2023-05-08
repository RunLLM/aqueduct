import argparse
import base64

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-s", "--spec", required=True)
    parser.add_argument("-v", "--version-tag", default="")
    args = parser.parse_args()
    if args.version_tag:
        import subprocess
        install_process = subprocess.run([
            "pip",
            "install",
            "-i",
            "https://test.pypi.org/simple/",
            f"aqueduct-ml={args.version_tag}",
        ])
        print(install_process.stderr)
        print(install_process.stdout)
        install_process.check_returncode()

    from aqueduct_executor.operators.airflow import execute
    from aqueduct_executor.operators.airflow.spec import parse_spec

    spec_json = base64.b64decode(args.spec)
    spec = parse_spec(spec_json)

    execute.run(spec, version_tag)
