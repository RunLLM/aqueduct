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

Requirements:
1) The script will wait for one successful run of any flow published by the notebook. It will infer that flow from the stdout
   of the notebook. Therefore, your notebook should do something akin to the following:
```
flow = client.publish_flow(...)
print(flow.id())
```