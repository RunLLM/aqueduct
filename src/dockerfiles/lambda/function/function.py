import base64
import subprocess
import sys


def pip_freeze(local_deps_path):
    subprocess.run([sys.executable, "-m", "pip", "freeze", ">>", local_deps_path])


def handler(event, context):
    """
    1. extract function
    2. download required packages
    3. execute function.
    """
    version_tag = event["VersionTag"]
    if version_tag:
        import subprocess
        install_process = subprocess.run([
            sys.executable,
            "-m",
            "pip",
            "install",
            "-i",
            "https://test.pypi.org/simple/",
            f"aqueduct-ml={version_tag}",
        ])
        print(install_process.stderr)
        print(install_process.stdout)
        install_process.check_returncode()

    input_spec = event["Spec"]

    from aqueduct_executor.operators.function_executor import (
        execute,
        extract_function,
        install_requirements,
    )
    from aqueduct_executor.operators.function_executor.spec import parse_spec
    spec_json = base64.b64decode(input_spec)
    spec = parse_spec(spec_json)

    extract_function.run(spec)
    open(spec.function_extract_path + "op/local_deps.txt", "w")
    open(spec.function_extract_path + "op/missing.txt", "w")
    pip_freeze(spec.function_extract_path + "op/local_deps.txt")
    install_requirements.run(
        spec.function_extract_path + "op/local_deps.txt",
        spec.function_extract_path + "op/requirements.txt",
        spec.function_extract_path + "op/missing.txt",
        spec,
    )
    execute.run(spec)
