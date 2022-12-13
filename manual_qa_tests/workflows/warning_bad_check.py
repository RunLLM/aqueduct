import aqueduct as aq

NAME = "warning_bad_check"
DESCRIPTION = """* Workflows Page: should succeed.
* Workflow Details Page: everything except `bad_check` artifact should succeed.
 `bad_check` artifact should fail with warning.
    * Workflow Status Bar: 0 error, 2 warning, 0 info, 6 success."""


@aq.metric(requirements=[])
def row_count(df):
    return df.shape[0]


@aq.check(requirements=[])
def good_check(count):
    return count > 10


@aq.check(requirements=[])
def bad_check(count):
    return count < 10


def deploy(client, integration_name):
    integration = client.integration(integration_name)
    reviews = integration.sql("SELECT * FROM hotel_reviews")
    row_count_artf = row_count(reviews)
    bad_check(row_count_artf)
    good_check(row_count_artf)
    client.publish_flow(
        artifacts=[reviews, row_count_artf],
        name=NAME,
        description=DESCRIPTION,
        schedule="",
    )
