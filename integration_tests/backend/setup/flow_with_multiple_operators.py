###
# Workflow that extracts a table, and simply passes it through
# four operators sequentially that simply return the table.
###
import aqueduct


def setup_flow_with_multiple_operators(
    client: aqueduct.Client,
    integration_name: str,
    workflow_name: str = "",
) -> str:
    name = workflow_name if workflow_name else "Test: Multiple Operators"
    n_runs = 1
    integration = client.resource(name=integration_name)

    @aqueduct.op
    def op1(df):
        return df

    @aqueduct.op
    def op2(df):
        return df

    @aqueduct.op
    def op3(df):
        return df

    @aqueduct.op
    def op4(df):
        return df

    reviews = integration.sql("SELECT * FROM hotel_reviews")
    df1 = op1(reviews)
    df2 = op2(df1)
    df3 = op3(df2)
    df4 = op4(df3)

    flow = client.publish_flow(artifacts=[df1, df2, df3, df4], name=name, description="Test description.")

    return flow.id(), n_runs
