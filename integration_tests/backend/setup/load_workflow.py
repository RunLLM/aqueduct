from typing import List

from aqueduct import Client, Flow


def workflow_load_tables_to_demo_db(
    client: Client, workflow_name: str, table_names: List[str], update_modes: List[str]
) -> Flow:
    integration = client.integration(name="aqueduct_demo")

    tables = []
    for table_name, update_mode in zip(table_names, update_modes):
        table = integration.sql(query="SELECT * FROM wine;")
        tables.append(table)
        table.save(integration.config(table=table_name, update_mode=update_mode))

    return client.publish_flow(name=workflow_name, artifacts=tables,)


def create_test_endpoint_GetWorkflowTables_flow(client, workflow_name, table_names, update_modes):
    for table_name, update_mode in zip(table_names, update_modes):
        flow = workflow_load_tables_to_demo_db(client, workflow_name, [table_name], [update_mode])
    return flow
