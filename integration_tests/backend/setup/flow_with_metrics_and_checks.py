###
# Workflow that extracts a table, and simply apply a row-count metric
# with a check to enforce the row-count is larger than 0.
# This workflow is published twice.
###
from aqueduct.constants.enums import CheckSeverity

import aqueduct


def setup_flow_with_metrics_and_checks(
    client: aqueduct.Client,
    resource_name: str,
    workflow_name: str = "",
) -> str:
    name = workflow_name if workflow_name else "Test: Flow with Metrics and Checks"
    n_runs = 2
    resource = client.resource(name=resource_name)

    @aqueduct.metric
    def size(df):
        return len(df)

    @aqueduct.check(severity=CheckSeverity.ERROR)
    def check(size):
        return size > 0

    reviews = resource.sql("SELECT * FROM hotel_reviews")
    rev_size = size(reviews)
    check_res = check(rev_size)

    flow = client.publish_flow(artifacts=[check_res], name=name)

    # publish again and triggers and update.
    rev_size = size(reviews)
    check_res = check(rev_size)
    flow = client.publish_flow(artifacts=[check_res], name=name)
    return flow.id(), n_runs
