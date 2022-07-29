import aqueduct

###
# Workflow that loads a table from the `aqueduct_demo` then saves it to `table_1` in append mode.
# This save operator is then replaced by one that saves to `table_1` in replace mode.
# In the next deployment of this run, it saves to `table_1` in append mode.
# In the last deployment, it saves to `table_2` in append mode.
###


def setup_changing_saves(client: aqueduct.Client, integration_name: str) -> str:
    name = "Test: Changing Saves"
    integration = client.integration(name=integration_name)

    ###

    table = integration.sql(query="SELECT * FROM wine;")

    table.save(integration.config(table="table_1", update_mode="append"))

    flow = client.publish_flow(
        name=name,
        artifacts=[table],
    )

    ###

    table = integration.sql(query="SELECT * FROM wine;")

    table.save(integration.config(table="table_1", update_mode="replace"))

    flow = client.publish_flow(
        name=name,
        artifacts=[table],
    )

    ###

    table = integration.sql(query="SELECT * FROM wine;")

    table.save(integration.config(table="table_1", update_mode="append"))

    flow = client.publish_flow(
        name=name,
        artifacts=[table],
    )

    ###

    table = integration.sql(query="SELECT * FROM wine;")

    table.save(integration.config(table="table_2", update_mode="append"))

    flow = client.publish_flow(
        name=name,
        artifacts=[table],
    )

    ###

    return flow.id()
