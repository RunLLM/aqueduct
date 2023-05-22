import aqueduct

###
# Workflow that loads a table from the `aqueduct_demo` then saves it to `table_1` in append mode.
# This save operator is then replaced by one that saves to `table_1` in replace mode.
# In the next deployment of this run, it saves to `table_1` in append mode.
# In the last deployment, it saves to `table_2` in append mode.
###


def setup_changing_saves(client: aqueduct.Client, resource_name: str) -> str:
    name = "Test: Changing Saves"
    n_runs = 4
    resource = client.resource(name=resource_name)

    ###
    table = resource.sql(query="SELECT * FROM wine;")
    resource.save(table, "table_1", "replace")
    flow = client.publish_flow(
        name=name,
        artifacts=[table],
    )

    ### update
    resource.save(table, "table_1", "append")
    flow = client.publish_flow(
        name=name,
        artifacts=[table],
    )

    ### update
    resource.save(table, "table_1", "append")
    flow = client.publish_flow(
        name=name,
        artifacts=[table],
    )

    ### update
    resource.save(table, "table_2", "replace")
    flow = client.publish_flow(
        name=name,
        artifacts=[table],
    )

    return flow.id(), n_runs
