import argparse
import base64

from aqueduct_executor.operators.system_metric_executor import execute
from aqueduct_executor.operators.system_metric_executor.spec import parse_spec

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-s", "--spec", required=True)
    args = parser.parse_args()

    spec_json = base64.b64decode(args.spec)
    spec = parse_spec(spec_json)

    execute.run(spec)
