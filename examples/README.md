The `run_notebook.py` is meant to run any notebook programmatically. It will not complete until
the published flow has made at least one successful run since the invocation of the notebook.

Example usage:
To run the notebook as is:
`python3 run_notebook.py --path "churn_prediction/Build and Deploy Churn Ensemble.ipynb"`

To run the notebook with a specific api key:
`python3 run_notebook.py --path "churn_prediction/Build and Deploy Churn Ensemble.ipynb" --api_key 97e385bad6eaee6ab6b082d0d1bfe2ba2c16bf23e7ec6e4f3f3fd00c27d5`

To run the notebook with a specific flow to wait for a successful run:
`python3 run_notebook.py --path "churn_prediction/Build and Deploy Churn Ensemble.ipynb" --flow_id ed1abcaf-fc35-4a73-90ae-51a726b8c757`

Requirements:
1) The credentials for the client initialized in the notebook must be formatted in the following fashion:
```
api_key = <api_key>
address = <server_address>
use_https = <True/False>
```

2) The script will wait for one successful run of any flow published by the notebook. It will infer that flow from the stdout
   of the notebook. Therefore, your notebook should do something akin to the following:
```
flow = client.publish_flow(...)
print(flow.id())
```