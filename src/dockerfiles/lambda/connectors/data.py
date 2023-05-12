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
            "--index-url",
            "https://test.pypi.org/simple/",
            "--extra-index-url",  # allows dependencies from pypi
            "https://pypi.org/simple",
            f"aqueduct-ml={version_tag}",
        ])
        print(install_process.stderr)
        print(install_process.stdout)
        install_process.check_returncode()

    input_spec = event["Spec"]

    from aqueduct_executor.operators.connectors.data import execute
    from aqueduct_executor.operators.connectors.data.spec import parse_spec
    
    spec_json = base64.b64decode(input_spec)
    spec = parse_spec(spec_json)

    execute.run(spec)
