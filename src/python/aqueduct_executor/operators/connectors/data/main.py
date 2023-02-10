import argparse
import base64
import time

from aqueduct_executor.operators.connectors.data import execute
from aqueduct_executor.operators.connectors.data.spec import parse_spec

if __name__ == "__main__":
    begin = time.time()

    parser = argparse.ArgumentParser()
    parser.add_argument("-s", "--spec", required=True)
    args = parser.parse_args()

    spec_json = base64.b64decode(args.spec)
    spec = parse_spec(spec_json)

    execute.run(spec)

    end = time.time()
    print("Connector took %s seconds." % (end - begin))
