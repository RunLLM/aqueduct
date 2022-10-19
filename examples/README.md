The `run_notebook.py` is meant to run any notebook programmatically. It will not complete until
the published flow has made at least one successful run since the invocation of the notebook.

Example usage:
To run the notebook as is:
`python3 run_notebook.py --path "churn_prediction/Quickstart Tutorial.ipynb"`

To run the notebook with a specific flow to wait for a successful run:
`python3 run_notebook.py --path "churn_prediction/Quickstart Tutorial.ipynb" --flow_id ed1abcaf-fc35-4a73-90ae-51a726b8c757`

To run the notebook with a specific server address, the address must be set in the notebook with the following format so the script can find and replace:
`address = <server_address>`. Then the command becomes:
`python3 run_notebook.py --path "churn_prediction/Quickstart Tutorial.ipynb --server_address=<...>`
