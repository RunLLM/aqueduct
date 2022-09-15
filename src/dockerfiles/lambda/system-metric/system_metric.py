import base64

from aqueduct_executor.operators.system_metric_executor import execute
from aqueduct_executor.operators.system_metric_executor.spec import parse_spec

def handler(event, context):
    print(event)
    input_spec = event["Spec"]

    spec_json = base64.b64decode(input_spec)
    spec = parse_spec(spec_json)

    execute.run(spec)
