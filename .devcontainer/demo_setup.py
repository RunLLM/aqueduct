import aqueduct as aq
from aqueduct.constants.enums import ExecutionStatus
import deploy_example
import time

def stop_condition(flow):
    # Last run is prepended to this list. We need to reverse it in order to compare with expected statuses,
    # which is sorted in chronologically ascending order.
    runs = list(reversed(flow.list_runs()))
    if len(runs) < 1:
        return False
    
    flow_run = runs[0]

    # Continue checking as long as there are still runs pending.
    if flow_run["status"] == ExecutionStatus.PENDING:
        return False
    return True

def polling(
    stop_condition_fn,
    timeout=60,
    poll_threshold=5,
    timeout_comment="Timed out waiting for workflow run to complete.",
):
    begin = time.time()

    while True:
        assert time.time() - begin < timeout, timeout_comment

        if stop_condition_fn():
            break
        else:
            time.sleep(poll_threshold)

client = aq.Client()
client.connect_integration("ncaa", "Postgres", {
    "host": "ec2-13-58-152-166.us-east-2.compute.amazonaws.com",
    "port": "5432",
    "database": "ncaa",
    "username": "aqueduct_demo",
    "password": "O1AspUC0UqIv2R6k",
})

deploy_example.deploy(
    "/notebook/", "demo.ipynb", "temp.py", "localhost:8080", aq.get_apikey()
)

flow = client.flow(flow_name="MarchMadnessWorkflow")

polling(
    lambda: stop_condition(flow),
    timeout=600,
    poll_threshold=10,
    timeout_comment="Timed out waiting for workflow run to complete.",
)
