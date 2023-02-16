import argparse
import base64

from aqueduct_executor.operators.function_executor import execute
from aqueduct_executor.operators.function_executor.spec import parse_spec
from aqueduct_executor.operators.utils.utils import time_it

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-s", "--spec", required=True)
    args = parser.parse_args()

    spec_json = base64.b64decode(args.spec)
    spec = parse_spec(spec_json)

    time_it(job_name=spec.name, job_type=spec.type.value, step="Running Operator (including IO)")(
        execute.run
    )(spec)
