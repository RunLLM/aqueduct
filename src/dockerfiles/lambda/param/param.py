import base64
import sys


def handler(event, context):
    print(event)
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

    from aqueduct_executor.operators.param_executor import execute
    from aqueduct_executor.operators.param_executor.spec import parse_spec
    
    spec_json = base64.b64decode(input_spec)
    spec = parse_spec(spec_json)

    execute.run(spec)
