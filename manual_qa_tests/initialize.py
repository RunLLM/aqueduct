import argparse

import deploy_example
from aqueduct.constants.enums import NotificationLevel
from notification import connect_slack
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

EXAMPLE_NOTEBOOKS_PATHS = [
    ["examples/churn_prediction/", "Customer Churn Prediction.ipynb"],
    ["examples/diabetes-classifier/", "Classifying Diabetes Risk.ipynb"],
    ["examples/house-price-prediction/", "House Price Prediction.ipynb"],
    ["examples/mpg-regressor/", "Predicting MPG.ipynb"],
    ["examples/sentiment-analysis/", "Sentiment Model.ipynb"],
    ["examples/wine-ratings-prediction/", "Predict Missing Wine Ratings.ipynb"],
]

TEMP_NOTEBOOK_PATH = "temp.py"
RUN_NOTEBOOK_SCRIPT = "examples/run_notebook.py"

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--addr", default="localhost:8080")
    parser.add_argument("--data-integration", default="aqueduct_demo")
    parser.add_argument("--api-key", default="")
    # parser.add_argument("-q", "--quiet", action="store_true")
    parser.add_argument("--example-notebooks", action="store_true")
    parser.add_argument("--slack-token", default="")
    parser.add_argument("--slack-channel", default="")
    parser.add_argument("--notification-level", default="success")
    args = parser.parse_args()

    api_key = args.api_key if args.api_key else aq.get_apikey()
    client = aq.Client(api_key, args.addr)

    if args.slack_token and args.slack_channel:
        connect_slack(
            client,
            args.slack_token,
            args.slack_channel,
            NotificationLevel(args.notification_level),
        )

    if args.example_notebooks:
        for example_path in EXAMPLE_NOTEBOOKS_PATHS:
            print(f"Deploying example notebooks {example_path[1]}...")
            deploy_example.deploy(
                example_path[0],
                example_path[1],
                TEMP_NOTEBOOK_PATH,
                args.addr,
                api_key,
            )

    for pkg in WORKFLOW_PKGS:
        print(f"Deploying {pkg.NAME}...")
        pkg.deploy(client, args.data_integration)
