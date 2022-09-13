import aqueduct

###
# Workflow that loads a table from the `aqueduct_demo` then saves it to `table_1` in append mode.
# This save operator is then replaced by one that saves to `table_1` in replace mode.
# In the next deployment of this run, it saves to `table_1` in append mode.
# In the last deployment, it saves to `table_2` in append mode.
###


def setup_flow_with_failure(client: aqueduct.Client, integration_name: str) -> str:
    name = "Test: Flow with Failure"
    n_runs = 1
    integration = client.integration(name=integration_name)

    @aqueduct.op()
    def bad_op(df):
        x = y  # intentional buggy code
        df["new"] = df["review"]

    @aqueduct.op()
    def bad_op_downstream(df):
        return df

    reviews = integration.sql("SELECT * FROM hotel_reviews")
    # use lazy mode to avoid previewing of bad_op
    # so that we can publish the flow
    bad_op_artf = bad_op.lazy(reviews)
    bad_op_downstream_artf = bad_op_downstream.lazy(bad_op_artf)

    flow = client.publish_flow(artifacts=[bad_op_downstream_artf], name=name)
    return flow.id(), n_runs
