import sys

api_key, server_address = sys.argv[1], sys.argv[2]

###
# Workflow that loads a table from the `aqueduct_demo` then saves it to `delete_table` in replace mode.
###

import aqueduct

name = "Test: Delete Workflow Tables"
client = aqueduct.Client(api_key, server_address)
integration = client.integration(name="aqueduct_demo")

###

table = integration.sql(query="SELECT * FROM wine;")

table.save(integration.config(table="delete_table", update_mode="append"))

flow = client.publish_flow(
    name=name,
    artifacts=[table],
)

###

print(flow.id())
