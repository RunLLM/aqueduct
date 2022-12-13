import aqueduct as aq

NAME = "succeed_parameters"
DESCRIPTION = """* Workflows Page: should fail.
* Workflow Details Page: should fail starting the first parameter.
    * Workflow Status Bar: 0 error, 0 warning, 0 info, 8 success.
    * Click into `table` parameters, sidesheet should also show failure.
    * There should be a older version that succeeded."""


@aq.check(requirements=[])
def check(df, bound):
    return df.shape[0] > bound


def deploy(client, integration_name):
    integration = client.integration(integration_name)
    client.create_param("table", default="hotel_reviews")
    bound = client.create_param("bound", default=10)

    reviews = integration.sql("SELECT * FROM {{ table }}")
    check(reviews, bound)
    client.publish_flow(
        artifacts=[reviews],
        name=NAME,
        description=DESCRIPTION,
        schedule="",
    )
