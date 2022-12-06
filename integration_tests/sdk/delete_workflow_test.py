import pandas as pd
import pytest
from aqueduct.error import InvalidRequestError
from constants import SHORT_SENTIMENT_SQL_QUERY
from utils import (
    check_flow_doesnt_exist,
    check_table_doesnt_exist,
    check_table_exists,
    generate_table_name,
    publish_flow_test,
    save,
)

from aqueduct import LoadUpdateMode


def test_delete_workflow_invalid_saved_objects(client, flow_name, data_integration, engine):
    """Check the flow cannot delete an object it had not saved."""
    integration = client.integration(data_integration)

    table = integration.sql(query=SHORT_SENTIMENT_SQL_QUERY)
    save(integration, table)

    flow = publish_flow_test(
        client,
        table,
        name=flow_name(),
        engine=engine,
    )

    tables = client.flow(flow.id()).list_saved_objects()
    tables[data_integration][0].object_name = "I_DON_T_EXIST"
    tables[data_integration] = [tables[data_integration][0]]

    # Cannot delete a flow if the saved objects specified had not been saved by the flow.
    with pytest.raises(InvalidRequestError):
        _ = client.delete_flow(flow.id(), saved_objects_to_delete=tables, force=True)

    # Check flow exists.
    client.flow(flow.id())


def test_delete_workflow_saved_objects(client, flow_name, data_integration, engine, validator):
    """Check the flow with object(s) saved with update_mode=APPEND can only be deleted if in force mode."""
    integration = client.integration(data_integration)
    table_name = generate_table_name()
    table = integration.sql(query=SHORT_SENTIMENT_SQL_QUERY)
    save(integration, table, name=table_name, update_mode=LoadUpdateMode.REPLACE)

    flow = publish_flow_test(
        client,
        table,
        name=flow_name(),
        engine=engine,
    )

    save(integration, table, name=table_name, update_mode=LoadUpdateMode.APPEND)
    flow = publish_flow_test(
        client,
        table,
        engine=engine,
        existing_flow=flow,
    )

    extracted_table_data = table.get()
    validator.check_saved_artifact(
        flow,
        table.id(),
        expected_data=pd.concat([extracted_table_data, extracted_table_data], ignore_index=True),
    )

    tables = client.flow(flow.id()).list_saved_objects()
    assert table_name in [item.object_name for item in tables[data_integration]]

    # Doesn't work if don't force because it is created in append mode.
    with pytest.raises(InvalidRequestError):
        client.delete_flow(flow.id(), saved_objects_to_delete=tables, force=False)

    # Actually delete the flow.
    client.delete_flow(flow.id(), saved_objects_to_delete=tables, force=True)
    check_flow_doesnt_exist(client, flow.id())

    # Check table no longer exists
    check_table_doesnt_exist(integration, table_name)


def test_delete_workflow_saved_objects_twice(
    client, flow_name, data_integration, engine, validator
):
    """Checking the successful deletion case and unsuccessful deletion case works as expected.
    To test this, I have two workflows that write to the same table. When I delete the table in the first workflow,
    it is successful but when I delete it in the second workflow, it is unsuccessful because the table has already
    been deleted.
    """
    integration = client.integration(data_integration)
    table_name = generate_table_name()

    table = integration.sql(query=SHORT_SENTIMENT_SQL_QUERY)
    save(integration, table, name=table_name, update_mode=LoadUpdateMode.REPLACE)

    # Workflow 1's name not specified, so given a random workflow name.
    flow1 = publish_flow_test(
        client,
        table,
        name=flow_name(),
        engine=engine,
    )

    # Workflow 2's name not specified, so given a random workflow name.
    save(integration, table, name=table_name, update_mode=LoadUpdateMode.APPEND)
    flow2 = publish_flow_test(
        client,
        table,
        name=flow_name(),
        engine=engine,
    )

    extracted_table_data = table.get()
    validator.check_saved_artifact(
        flow1,
        table.id(),
        expected_data=pd.concat([extracted_table_data, extracted_table_data], ignore_index=True),
    )

    # Check table is properly created at the integration.
    # Need to poll initially in case still writing table.
    check_table_exists(integration, table_name)

    tables = client.flow(flow1.id()).list_saved_objects()
    tables_1 = set([item.object_name for item in tables[data_integration]])
    assert table_name in tables_1

    tables = client.flow(flow2.id()).list_saved_objects()
    tables_2 = set([item.object_name for item in tables[data_integration]])
    assert table_name in tables_2

    assert tables_1 == tables_2
    client.delete_flow(flow1.id(), saved_objects_to_delete=tables, force=True)

    # Check flow indeed deleted and that the table no longer exists.
    check_flow_doesnt_exist(client, flow1.id())
    check_table_doesnt_exist(integration, table_name)

    # Try to delete table deleted by other flow.
    with pytest.raises(Exception) as e_info:
        client.delete_flow(flow2.id(), saved_objects_to_delete=tables, force=True)
    assert str(e_info.value).startswith("Failed to delete")

    # Failed to delete tables, but flow should be removed.
    check_flow_doesnt_exist(client, flow2.id())
