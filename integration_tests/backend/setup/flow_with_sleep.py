import time

import aqueduct

"""
`setup_flow_with_sleep` sets up a workflow like the following:
sleeping_op() (sleeps for 60s before succeed)

When using this flow for testing, the caller need to ensure
the delay retrieving its status is less than 60s.
"""


def setup_flow_with_sleep(client: aqueduct.Client, integration_name: str) -> str:
    name = "Test: Flow with Sleep"
    n_runs = 1
    integration = client.integration(name=integration_name)

    @aqueduct.op
    def sleeping_op(df):
        time.sleep(60)
        return df

    reviews = integration.sql("SELECT * FROM hotel_reviews")
    # use lazy mode to avoid previewing of bad_op
    # so that we can publish the flow
    sleeping_artf = sleeping_op.lazy(reviews)

    flow = client.publish_flow(artifacts=[sleeping_artf], name=name)
    return flow.id(), n_runs
