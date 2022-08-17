import pytest
from aqueduct.error import InvalidRequestError, InvalidUserArgumentException
from aqueduct.error import AqueductError
from constants import SHORT_SENTIMENT_SQL_QUERY
from utils import (
    delete_flow,
    generate_new_flow_name,
    get_integration_name,
    run_flow_test,
    polling,
    check_flow_doesnt_exist,
    check_table_exists,
    check_table_doesnt_exist,
)


from aqueduct import LoadUpdateMode

LIST_INTEGRATION_OBJECTS_TEMPLATE = "/api/integration/%s/objects"

# Check the flow cannot delete an object it had not saved.
@pytest.mark.publish
def test_delete_workflow_invalid_saved_objects(client):
    integration = client.integration(name=get_integration_name())
    name = generate_new_flow_name()
    table_name = generate_new_flow_name()
    flow_id = None

    ###

    table = integration.sql(query=SHORT_SENTIMENT_SQL_QUERY)

    table.save(integration.config(table=table_name, update_mode=LoadUpdateMode.REPLACE))

    flow_id = run_flow_test(client, [table], name=name, num_runs=1, delete_flow_after=False).id()

    ###

    try:
        tables = client.flow(flow_id).list_saved_objects()
        tables[get_integration_name()][0].object_name = "I_DON_T_EXIST"
        tables[get_integration_name()] = [tables[get_integration_name()][0]]

        # Cannot delete a flow if the saved objects specified had not been saved by the flow.
        with pytest.raises(InvalidRequestError):
            data = client.delete_flow(flow_id, saved_objects_to_delete=tables, force=True)
        
        # Check flow exists.
        client.flow(flow_id)
    finally:
        delete_flow(client, flow_id)

# Check the flow with object(s) saved with update_mode=APPEND can only be deleted if in force mode.
@pytest.mark.publish
def test_delete_workflow_saved_objects(client):
    integration = client.integration(name=get_integration_name())
    name = generate_new_flow_name()
    table_name = generate_new_flow_name()
    flow_ids_to_delete = set()
    endpoint = LIST_INTEGRATION_OBJECTS_TEMPLATE % integration._metadata.id

    ###

    table = integration.sql(query=SHORT_SENTIMENT_SQL_QUERY)

    table.save(integration.config(table=table_name, update_mode=LoadUpdateMode.REPLACE))

    flow_ids_to_delete.add(
        run_flow_test(client, [table], name=name, num_runs=1, delete_flow_after=False).id()
    )

    ###

    table.save(integration.config(table=table_name, update_mode=LoadUpdateMode.APPEND))

    flow_ids_to_delete.add(
        run_flow_test(client, [table], name=name, num_runs=2, delete_flow_after=False).id()
    )

    ###

    try:
        assert len(flow_ids_to_delete) == 1
        flow_id = list(flow_ids_to_delete)[0]
        tables = client.flow(flow_id).list_saved_objects()

        assert table_name in [item.object_name for item in tables[get_integration_name()]]

        # Check table is properly created at the integration.
        # Need to poll initially in case still writing table.
        check_table_exists(integration, table_name)

        # Doesn't work if don't force because it is created in append mode.
        with pytest.raises(InvalidRequestError):
            client.delete_flow(flow_id, saved_objects_to_delete=tables, force=False)

        # Check table is properly created at the integration.
        integration.sql(f"SELECT * FROM {table_name}").get()

        client.delete_flow(flow_id, saved_objects_to_delete=tables, force=True)
        
        # Check flow indeed deleted
        check_flow_doesnt_exist(client, flow_id)
        flow_ids_to_delete.remove(flow_id)

        # Check table no longer exists
        check_table_doesnt_exist(integration, table_name)

    finally:
        for flow_id in flow_ids_to_delete:
            delete_flow(client, flow_id)

# Checking the successful deletion case and unsuccessful deletion case works as expected.
# To test this, I have two workflows that write to the same table. When I delete the table in the first workflow, 
# it is successful but when I delete it in the second workflow, it is unsuccessful because the table has already 
# been deleted.
@pytest.mark.publish
def test_delete_workflow_saved_objects_twice(client):
    integration = client.integration(name=get_integration_name())
    name = generate_new_flow_name()
    table_name = generate_new_flow_name()
    flow_ids_to_delete = set()
    endpoint = LIST_INTEGRATION_OBJECTS_TEMPLATE % integration._metadata.id

    ###

    table = integration.sql(query=SHORT_SENTIMENT_SQL_QUERY)

    table.save(integration.config(table=table_name, update_mode=LoadUpdateMode.REPLACE))

    # Workflow 1's name not specified, so given a random workflow name.
    flow_ids_to_delete.add(run_flow_test(client, [table], num_runs=1, delete_flow_after=False).id())

    ###

    table.save(integration.config(table=table_name, update_mode=LoadUpdateMode.APPEND))

    # Workflow 2's name not specified, so given a random workflow name.
    flow_ids_to_delete.add(run_flow_test(client, [table], num_runs=1, delete_flow_after=False).id())

    ###

    try:
        assert len(flow_ids_to_delete) == 2
        flow_list = list(flow_ids_to_delete)
        flow_1_id = flow_list[0]
        flow_2_id = flow_list[1]

        # Check table is properly created at the integration.
        # Need to poll initially in case still writing table.
        check_table_exists(integration, table_name)

        tables = client.flow(flow_1_id).list_saved_objects()
        tables_1 = set([item.object_name for item in tables[get_integration_name()]])
        assert table_name in tables_1

        tables = client.flow(flow_2_id).list_saved_objects()
        tables_2 = set([item.object_name for item in tables[get_integration_name()]])
        assert table_name in tables_2

        assert tables_1 == tables_2

        client.delete_flow(flow_1_id, saved_objects_to_delete=tables, force=True)

        # Check flow indeed deleted
        check_flow_doesnt_exist(client, flow_1_id)
        flow_ids_to_delete.remove(flow_1_id)


        # Check table no longer exists
        check_table_doesnt_exist(integration, table_name)

        # Try to delete table deleted by other flow.
        with pytest.raises(Exception) as e_info:
            client.delete_flow(flow_2_id, saved_objects_to_delete=tables, force=True)
        assert str(e_info.value).startswith("Failed to delete")


        # Failed to delete tables, but flow should be removed.
        check_flow_doesnt_exist(client, flow_2_id)
        flow_ids_to_delete.remove(flow_2_id)

    finally:
        for flow_id in flow_ids_to_delete:
            delete_flow(client, flow_id)
