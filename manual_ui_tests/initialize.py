import argparse

from workflows import fail_bad_check, succeed_complex, succeed_parameters, warning_bad_check

import aqueduct as aq

# when adding new deployments, keep the order of `fail`, `warning`, and `succeed`
# such that the UI would approximately show these workflows in reverse order.
WORKFLOW_PKGS = [
    fail_bad_check,
    warning_bad_check,
    succeed_parameters,
    succeed_complex,
]

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--addr", default="localhost:8080")
    parser.add_argument("--integration", default="aqueduct_demo")
    parser.add_argument("--api-key", default="")
    parser.add_argument("-q", "--quiet", action="store_true")
    args = parser.parse_args()

    api_key = args.api_key if args.api_key else aq.get_apikey()
    client = aq.Client(api_key, args.addr)

    for pkg in WORKFLOW_PKGS:
        if not args.quiet:
            print(f"Deploying {pkg.NAME}...")
        pkg.deploy(client, args.integration)
