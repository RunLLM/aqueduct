import argparse
import base64
import json
import time

from aqueduct_executor.operators.function_executor import execute
from aqueduct_executor.operators.function_executor.spec import parse_spec
from aqueduct_executor.operators.utils.utils import print_with_color

if __name__ == "__main__":
    begin = time.time()

    parser = argparse.ArgumentParser()
    parser.add_argument("-s", "--spec", required=True)
    args = parser.parse_args()

    spec_json = base64.b64decode(args.spec)
    spec = parse_spec(spec_json)

    execute.run(spec)

    end = time.time()
    performance = {
        "job": spec.name,
        "step": "Running Operator (including IO)",
        "latency(s)": (end - begin),
    }
    print_with_color(json.dumps(performance, indent=4))
