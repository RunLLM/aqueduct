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