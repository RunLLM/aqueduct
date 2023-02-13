import argparse
import base64
import json
import time

from aqueduct_executor.operators.connectors.data import execute
from aqueduct_executor.operators.connectors.data.spec import parse_spec
from aqueduct_executor.operators.utils.enums import PrintColorType
from aqueduct_executor.operators.utils.utils import print_with_color

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-s", "--spec", required=True)
    args = parser.parse_args()

    spec_json = base64.b64decode(args.spec)
    spec = parse_spec(spec_json)

    print_with_color(
        "Starting %s job: %s" % (spec.type.value, spec.name), color=PrintColorType.GREEN
    )
    begin = time.time()

    execute.run(spec)

    end = time.time()
    performance = {
        "job": spec.name,
        "type": spec.type,
        "step": "Running Connector",
        "latency(s)": (end - begin),
    }
    print_with_color(json.dumps(performance, indent=4))
