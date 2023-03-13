import aqueduct as aq
from aqueduct.constants.enums import ExecutionStatus
import time

NAME = "fail_bad_op"
DESCRIPTION = """* Workflows Page: should fail.
* Workflow Details Page:
    * Everything before `bad_op` should succeed.
    * `bad_op` should fail.
    * Everything after `bad_op` should be canceled.
    * Metric and check details should show a list of canceled history, but not a plot.
"""


@aq.metric(requirements=[])
def row_count(df):
    return df.shape[0]


@aq.check(requirements=[], severity=aq.constants.enums.CheckSeverity.ERROR)
def check(count):
    return count < 10


@aq.op(requirements=[])
def bad_op(_):
    x = [1]
    return x[2]


def deploy(client, integration_name):
    integration = client.integration(integration_name)
    reviews = integration.sql("SELECT * FROM hotel_reviews")
    bad_op_artf = bad_op.lazy(reviews)
    row_count_artf = row_count.lazy(bad_op_artf)
    # using lazy() to bypass preview
    check_artf = check.lazy(row_count_artf)
    flow = client.publish_flow(
        artifacts=[check_artf],
        name=NAME,
        description=DESCRIPTION,
        schedule="",
    )

    polling(
        lambda: stop_condition(flow),
        timeout=600,
        poll_threshold=10,
        timeout_comment="Timed out waiting for workflow run to complete.",
    )

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
