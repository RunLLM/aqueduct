import aqueduct as aq

NAME = "succeed_with_snapshots_disabled"
DESCRIPTION = """* Workflows Page: should succeed.
* Workflow Details Page:
  * There artifacts are shown deleted: op_disabled artf, Demo query artf
  * Both checks should show 'passed'.
"""


@aq.op(requirements=[])
def op_disabled(df):
    return df


@aq.metric(requirements=[])
def metric(df):
    return df.shape[0]


@aq.check(requirements=[])
def check(count):
    return count > 10


@aq.check(requirements=[])
def check_disabled(count):
    return count > 10


def deploy(client, resource_name):
    resource = client.resource(resource_name)
    reviews = resource.sql("SELECT * FROM hotel_reviews")
    op_artf = op_disabled(reviews)
    metric_artf = metric(op_artf)
    check(metric_artf)
    check_disabled_artf = check_disabled(metric_artf)

    op_artf.enable_snapshot()
    check_disabled_artf.disable_snapshot()

    resource.save(op_artf, "succeed_with_snapshots_disabled_tmp", aq.LoadUpdateMode.REPLACE)

    client.publish_flow(
        artifacts=[op_artf],
        name=NAME,
        description=DESCRIPTION,
        schedule="",
        disable_snapshots=True,
    )
