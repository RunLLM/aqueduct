import argparse
import base64

from aqueduct_executor.migrators.parameter_val_type_inference_000019 import execute
from aqueduct_executor.migrators.parameter_val_type_inference_000019.spec import parse_spec

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-s", "--spec", required=True)
    args = parser.parse_args()

    spec_json = base64.b64decode(args.spec)
    spec = parse_spec(spec_json)
    if spec.op == "encode":
        execute.run_type_inference_and_encode(spec.param_val)
    else:
        execute.run_decode(spec.param_val, spec.param_type)
