import base64

from aqueduct_executor.operators.system_metric_executor import execute
from aqueduct_executor.operators.system_metric_executor.spec import parse_spec


def handler(event, context):
    print(event)
    version_tag = event["VersionTag"]
    if version_tag:
        import subprocess
        install_process = subprocess.run([
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

    spec_json = base64.b64decode(input_spec)
    spec = parse_spec(spec_json)

    execute.run(spec)
