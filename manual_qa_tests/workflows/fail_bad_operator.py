import aqueduct as aq

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
    client.publish_flow(
        artifacts=[check_artf],
        name=NAME,
        description=DESCRIPTION,
        schedule="",
    )
