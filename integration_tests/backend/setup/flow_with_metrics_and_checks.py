import aqueduct

###
# Workflow that loads a table from the `aqueduct_demo` then saves it to `table_1` in append mode.
# This save operator is then replaced by one that saves to `table_1` in replace mode.
# In the next deployment of this run, it saves to `table_1` in append mode.
# In the last deployment, it saves to `table_2` in append mode.
###


def setup_flow_with_metrics_and_checks(client: aqueduct.Client, integration_name: str) -> str:
    name = "Test: Flow with Metrics and Bad Checks"
    n_runs = 1
    integration = client.integration(name=integration_name)

    @aqueduct.metric
    def size(df):
        return len(df)

    @aqueduct.check(severity=aqueduct.CheckSeverity.ERROR)
    def check(size):
        return size > 0

    reviews = integration.sql("SELECT * FROM hotel_reviews")
    rev_size = size(reviews)
    check_res = check(rev_size)

    flow = client.publish_flow(artifacts=[check_res], name=name)
    return flow.id(), n_runs
