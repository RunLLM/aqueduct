import time

import aqueduct as aq

NAME = "succeed_parameters"
DESCRIPTION = """* Workflows Page: should succeed.
* Workflow Details Page: everything should be green.
    * There should be 2 versions.
    * Click into `bound` parameters, the value of the later version should be 5.
    The value of the older version should be 10.
    * Click into `empty_str` parameter. Nothing should show underneath Preview header."""


@aq.check(requirements=[])
def check(df, bound):
    return df.shape[0] > bound

@aq.check(requirements=[])
def empty_str_check(df, empty_str):
    return len(empty_str) == 0

def deploy(client, integration_name):
    integration = client.integration(integration_name)
    client.create_param("table", default="hotel_reviews")
    bound = client.create_param("bound", default=10)
    empty_str = client.create_param("empty_str", default="")

    reviews = integration.sql("SELECT * FROM {{ table }}")
    check(reviews, bound)
    empty_str_check(reviews, empty_str)
    flow = client.publish_flow(
        artifacts=[reviews],
        name=NAME,
        description=DESCRIPTION,
        schedule="",
    )
    time.sleep(5)
    client.trigger(flow.id(), parameters={"bound": 5, "empty_str": ""})
