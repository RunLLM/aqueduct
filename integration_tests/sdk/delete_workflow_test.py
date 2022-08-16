from time import sleep, time

import pytest
from aqueduct.error import InvalidRequestError
from aqueduct.error import AqueductError
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
        tables[get_integration_name()][0].object_name = "I_DON_T_EXIST"
        tables[get_integration_name()] = [tables[get_integration_name()][0]]

        with pytest.raises(InvalidRequestError):
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
        run_flow_test(client, [table], name=name, num_runs=2, delete_flow_after=False).id()
    )

    ###

    try:
        assert len(flow_ids_to_delete) == 1
        flow_id = list(flow_ids_to_delete)[0]
        tables = client.flow(flow_id).list_saved_objects()

        assert "delete_table" in [item.object_name for item in tables[get_integration_name()]]

        # No SDK function to do this so we query the endpoint directly to see delete_table is properly created at the integration.
        tables_response = get_response(client, endpoint).json()
        assert "delete_table" in set(tables_response["object_names"])

        # Doesn't work if don't force
        with pytest.raises(InvalidRequestError):
            client.delete_flow(flow_id, saved_objects_to_delete=tables, force=False)

        # Wait for deletion to occur
        sleep(1)

        # No SDK function to do this so we query the endpoint directly to see delete_table is properly deleted at the integration.
        tables_response = get_response(client, endpoint).json()
        assert "delete_table" in set(tables_response["object_names"])

        client.delete_flow(flow_id, saved_objects_to_delete=tables, force=True)

        flow_ids_to_delete.remove(flow_id)

        # Wait for deletion to occur
        sleep(1)

        # No SDK function to do this so we query the endpoint directly to see delete_table is properly deleted at the integration.
        tables_response = get_response(client, endpoint).json()
        assert "delete_table" not in set(tables_response["object_names"])

    finally:
        for flow_id in flow_ids_to_delete:
            delete_flow(client, flow_id)


@pytest.mark.publish
def test_delete_workflow_saved_objects_twice(client):
    integration = client.integration(name=get_integration_name())
    name = generate_new_flow_name()
    flow_ids_to_delete = set()
    endpoint = LIST_INTEGRATION_OBJECTS_TEMPLATE % integration._metadata.id

    ###

    table = integration.sql(query=SENTIMENT_SQL_QUERY)

    table.save(integration.config(table="delete_table", update_mode=LoadUpdateMode.REPLACE))

    flow_ids_to_delete.add(run_flow_test(client, [table], num_runs=1, delete_flow_after=False).id())

    ###

    table.save(integration.config(table="delete_table", update_mode=LoadUpdateMode.APPEND))

    flow_ids_to_delete.add(run_flow_test(client, [table], num_runs=1, delete_flow_after=False).id())

    ###

    try:
        assert len(flow_ids_to_delete) == 2
        flow_list = list(flow_ids_to_delete)
        flow_1_id = flow_list[0]
        flow_2_id = flow_list[1]

        tables = client.flow(flow_1_id).list_saved_objects()
        tables_1 = set([item.object_name for item in tables[get_integration_name()]])
        assert "delete_table" in tables_1

        tables = client.flow(flow_2_id).list_saved_objects()
        tables_2 = set([item.object_name for item in tables[get_integration_name()]])
        assert "delete_table" in tables_2

        assert tables_1 == tables_2

        # Check delete_table is properly created at the integration.
        integration.sql("SELECT * FROM delete_table").get()

        client.delete_flow(flow_1_id, saved_objects_to_delete=tables, force=True)

        flow_ids_to_delete.remove(flow_1_id)

        # Wait for deletion to occur
        timeout = 500
        poll_threshold = 5
        begin = time()

        while True:
            assert time() - begin < timeout, "Timed out waiting for workflow run to complete."

            # Check delete_table no longer exists.
            try:
                integration.sql("SELECT * FROM delete_table").get()
                sleep(poll_threshold)
            except:
                break

        # Try to delete table deleted by other flow.
        with pytest.raises(Exception) as e_info:
            client.delete_flow(flow_2_id, saved_objects_to_delete=tables, force=True)
        assert str(e_info.value).startswith("Failed to delete")

        # Failed to delete tables, but flow should be removed.
        with pytest.raises(Exception) as e_info:
            client.delete_flow(flow_2_id)
        assert str(e_info.value).startswith("The organization does not own")
        flow_ids_to_delete.remove(flow_2_id)

    finally:
        for flow_id in flow_ids_to_delete:
            delete_flow(client, flow_id)
