import argparse
import base64

<<<<<<< HEAD:src/python/aqueduct_executor/operators/connectors/data/main.py
from aqueduct_executor.operators.connectors.data import execute
from aqueduct_executor.operators.connectors.data.spec import parse_spec
=======
from aqueduct_executor.migrators.artifact_migration import execute
from aqueduct_executor.migrators.artifact_migration.spec import parse_spec
>>>>>>> 1b040880a5301159d3cbf0c3344f992bdce46744:src/python/aqueduct_executor/migrators/artifact_migration/main.py

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-s", "--spec", required=True)
    args = parser.parse_args()

    spec_json = base64.b64decode(args.spec)
    spec = parse_spec(spec_json)

    execute.run(spec)
