import aqueduct as aq

NAME = "fail_bad_check"
DESCRIPTION = """* Workflows Page: should fail.
* Workflow Details Page: everything except `bad_check` artifact should succeed.
 `bad_check` artifact should fail.
    * Workflow Status Bar: 2 error, 0 warning, 0 info, 4 success."""


@aq.metric(requirements=[])
def row_count(df):
    return df.shape[0]


@aq.check(requirements=[], severity=aq.constants.enums.CheckSeverity.ERROR)
def bad_check(count):
    return count < 10


def deploy(client, integration_name):
    integration = client.integration(integration_name)
    reviews = integration.sql("SELECT * FROM hotel_reviews")
    row_count_artf = row_count(reviews)
    # using lazy() to bypass preview
    bad_check_artf = bad_check.lazy(row_count_artf)
    client.publish_flow(
        artifacts=[reviews, row_count_artf, bad_check_artf],
        name=NAME,
        description=DESCRIPTION,
        schedule="",
    )
