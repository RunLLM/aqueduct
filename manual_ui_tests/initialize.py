import argparse
import aqueduct as aq
from workflows import (
    fail_bad_check,
    should_succeed,
)

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--addr", default="localhost:8080")
    parser.add_argument("--integration", default="aqueduct_demo")
    parser.add_argument("--api-key", default="")
    parser.add_argument("--verbose")
    args = parser.parse_args()

    api_key = args.api_key if args.api_key else aq.get_apikey()
    client = aq.Client(api_key, args.addr)

    should_succeed.deploy(client, args.integration)
    fail_bad_check.deploy(client, args.integration)