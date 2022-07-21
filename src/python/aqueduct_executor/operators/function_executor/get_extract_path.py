import argparse
import base64

from aqueduct_executor.operators.function_executor.spec import FunctionSpec, parse_spec


def run(spec: FunctionSpec) -> str:
    return spec.function_extract_path


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-s", "--spec", required=True)
    args = parser.parse_args()

    spec_json = base64.b64decode(args.spec)
    spec = parse_spec(spec_json)
    # The output of the print statement to stdout is captured by the calling bash script into a variable,
    # so we should not include any other print statements in this Python script.
    print(run(spec))
