import argparse
import deploy_example

from aqueduct.constants.enums import NotificationLevel
from multiprocessing import Process
from notification import connect_slack
from wait_for_flows import wait_for_all_flows_to_complete
from workflows import (
    check_status_test,
    fail_bad_check,
    fail_bad_operator,
    no_run,
    succeed_complex,
    succeed_dag_layout_test,
    succeed_march_madness_dag_layout_test,
    succeed_parameters,
    warning_bad_check,
)

from workflows.check_status_test import hello

import aqueduct as aq

# when adding new deployments, keep the order of `fail`, `warning`, and `succeed`
# such that the UI would approximately show these workflows in reverse order.
WORKFLOW_PKGS = [
    check_status_test,
    fail_bad_check,
    warning_bad_check,
    succeed_parameters,
    succeed_complex,
    succeed_dag_layout_test,
    succeed_march_madness_dag_layout_test,
    fail_bad_operator,
    no_run,
]

DEMO_NOTEBOOKS_PATHS = [
    ["examples/wine-ratings-prediction/", "Predict Missing Wine Ratings.ipynb"],
    ["examples/churn_prediction/", "Customer Churn Prediction.ipynb"],
    ["examples/diabetes-classifier/", "Classifying Diabetes Risk.ipynb"],
    ["examples/house-price-prediction/", "House Price Prediction.ipynb"],
    ["examples/mpg-regressor/", "Predicting MPG.ipynb"],
]

ADDITIONAL_EXAMPLE_NOTEBOOKS_PATHS = [
    ["examples/sentiment-analysis/", "Sentiment Model.ipynb"],
]

TEMP_NOTEBOOK_PATH = "temp.py"
RUN_NOTEBOOK_SCRIPT = "examples/run_notebook.py"


def deploy_flow(name, deploy_fn, api_key, address, data_integration) -> None:
    print(f"Deploying {name}...")
    client = aq.Client(api_key, address)
    deploy_fn(client, data_integration)

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--addr", default="localhost:8080")
    parser.add_argument("--data-integration", default="aqueduct_demo")
    parser.add_argument("--api-key", default="")
    # parser.add_argument("-q", "--quiet", action="store_true")
    parser.add_argument("--example-notebooks", action="store_true")
    parser.add_argument("--example-notebooks-only", action="store_true")
    parser.add_argument("--demo-container-notebooks-only", action="store_true")
    parser.add_argument("--slack-token", default="")
    parser.add_argument("--slack-channel", default="")
    parser.add_argument("--notification-level", default="success")
    parser.add_argument("--wait-to-complete", action="store_true")
    args = parser.parse_args()

    api_key = args.api_key if args.api_key else aq.get_apikey()

    if args.slack_token and args.slack_channel:
        connect_slack(
            client,
            args.slack_token,
            args.slack_channel,
            NotificationLevel(args.notification_level),
        )

    if args.example_notebooks or args.example_notebooks_only or args.demo_container_notebooks_only:
        notebooks = DEMO_NOTEBOOKS_PATHS
        if not args.demo_container_notebooks_only:
            notebooks += ADDITIONAL_EXAMPLE_NOTEBOOKS_PATHS

        for example_path in notebooks:
            print(f"Deploying example notebooks {example_path[1]}...")
            deploy_example.deploy(
                example_path[0],
                example_path[1],
                TEMP_NOTEBOOK_PATH,
                args.addr,
                api_key,
            )

    if not args.example_notebooks_only and not args.demo_container_notebooks_only:
        processes = []
        for pkg in WORKFLOW_PKGS:
            p = Process(target=deploy_flow, args=(pkg.NAME, pkg.deploy, api_key, args.addr, args.data_integration))
            processes.append(p)
            p.start()

        for p in processes:
            p.join()

    if args.wait_to_complete:
        wait_for_all_flows_to_complete(client)
