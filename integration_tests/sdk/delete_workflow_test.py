import pytest
from constants import SENTIMENT_SQL_QUERY
from utils import (
    delete_flow,
    generate_new_flow_name,
    get_integration_name,
    get_response,
    run_flow_test,
)

from aqueduct import LoadUpdateMode

LIST_INTEGRATION_OBJECTS_TEMPLATE = "/api/integration/%s/objects"


@pytest.mark.publish
def test_delete_workflow_invalid_saved_objects(client):
    integration = client.integration(name=get_integration_name())
    name = generate_new_flow_name()
    flow_id = None

    ###

    table = integration.sql(query=SENTIMENT_SQL_QUERY)

    table.save(integration.config(table="delete_table", update_mode=LoadUpdateMode.REPLACE))

    flow_id = run_flow_test(client, [table], name=name, num_runs=1, delete_flow_after=False).id()

    ###

    try:
        tables = client.flow(flow_id).list_saved_objects()
        tables[get_integration_name()][0].name = "I_DON_T_EXIST"
        tables[get_integration_name()] = [tables[get_integration_name()][0]]

        with pytest.raises(InvalidRequestError) as e_info:
            data = client.delete_flow(flow_id, saved_objects_to_delete=tables, force=True)
    finally:
        delete_flow(client, flow_id)


@pytest.mark.publish
def test_delete_workflow_saved_objects(client):
    integration = client.integration(name=get_integration_name())
    name = generate_new_flow_name()
    flow_ids_to_delete = set()
    endpoint = LIST_INTEGRATION_OBJECTS_TEMPLATE % integration._metadata.id

    ###

    table = integration.sql(query=SENTIMENT_SQL_QUERY)

    table.save(integration.config(table="delete_table", update_mode=LoadUpdateMode.REPLACE))

    flow_ids_to_delete.add(
        run_flow_test(client, [table], name=name, num_runs=1, delete_flow_after=False).id()
    )

    ###

    table.save(integration.config(table="delete_table", update_mode=LoadUpdateMode.APPEND))

    flow_ids_to_delete.add(
        run_flow_test(client, [table], name=name, num_runs=1, delete_flow_after=False).id()
    )

    ###

    try:
        assert len(flow_ids_to_delete) == 1
        flow_id = list(flow_ids_to_delete)[0]
        tables = client.flow(flow_id).list_saved_objects()

        assert "delete_table" in tables[get_integration_name()]

        tables_response = get_response(client, endpoint).json()

        assert "delete_table" in set(tables_response["object_names"])

        with pytest.raises(InvalidRequestError) as e_info:
            client.delete_flow(flow_id, saved_objects_to_delete=tables, force=False)

        client.delete_flow(flow_id, saved_objects_to_delete=tables, force=True)

        # Wait for deletion to occur
        sleep(1)

        tables_response = get_response(client, endpoint).json()

        assert "delete_table" not in set(tables_response["object_names"])

    finally:
        for flow_id in flow_ids_to_delete:
            delete_flow(client, flow_id)
