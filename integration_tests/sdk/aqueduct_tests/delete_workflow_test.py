import pandas as pd
import pytest
from aqueduct.constants.enums import LoadUpdateMode
from aqueduct.error import InvalidRequestError, InvalidUserArgumentException

import aqueduct
from aqueduct import op

from ..shared.data_objects import DataObject
from ..shared.flow_helpers import publish_flow_test
from ..shared.naming import generate_table_name
from ..shared.relational import all_relational_DBs
from .extract import extract
from .save import save


def test_delete_source_workflow(client, flow_name, data_resource, engine):
    """Tests deleting a flow that is the source of another flow that has a cascading trigger."""
    table_artifact = extract(data_resource, DataObject.SENTIMENT)

    @op
    def noop(input):
        return input

    output_artifact = noop(table_artifact)

    # Create a source flow
    source_name = flow_name()
    source_flow = publish_flow_test(
        client,
        name=source_name,
        artifacts=output_artifact,
        engine=engine,
        schedule=aqueduct.daily(),
    )

    # Create a flow that is set to run after the above source_flow
    name = flow_name()
    flow = publish_flow_test(
        client,
        name=name,
        artifacts=output_artifact,
        engine=engine,
        source_flow=source_flow,
    )

    # Delete source flow
    client.delete_flow(source_flow.id())


def test_delete_workflow_invalid_saved_objects(client, flow_name, data_resource, engine):
    """Check the flow cannot delete an object it had not saved."""
    table = extract(data_resource, DataObject.SENTIMENT)
    save(data_resource, table)

    flow = publish_flow_test(
        client,
        table,
        name=flow_name(),
        engine=engine,
    )

    tables = client.flow(flow.id()).list_saved_objects()
    table_saved_object_update = tables[data_resource][0]
    table_saved_object_update.spec.set_identifier("I_DONT_EXIST")
    tables[data_resource] = [table_saved_object_update]

    # Cannot delete a flow if the saved objects specified had not been saved by the flow.
    with pytest.raises(InvalidRequestError):
        client.delete_flow(flow.id(), saved_objects_to_delete=tables, force=True)

    # Check flow exists.
    client.flow(flow.id())


@pytest.mark.enable_only_for_data_resource_type(*all_relational_DBs())
# TODO(ENG-2881)
@pytest.mark.skip_for_spark_engines(reason="Need to investigate, TODO above.")
def test_force_delete_workflow_saved_objects(
    client, flow_name, data_resource, engine, data_validator
):
    """Check the flow with object(s) saved with update_mode=APPEND can only be deleted if in force mode."""
    table_name = generate_table_name()
    table_artifact = data_resource.sql("select * from hotel_reviews limit 5")
    save(data_resource, table_artifact, name=table_name, update_mode=LoadUpdateMode.REPLACE)

    flow = publish_flow_test(
        client,
        table_artifact,
        name=flow_name(),
        engine=engine,
    )

    save(data_resource, table_artifact, name=table_name, update_mode=LoadUpdateMode.APPEND)
    flow = publish_flow_test(
        client,
        table_artifact,
        engine=engine,
        existing_flow=flow,
    )

    extracted_table_data = table_artifact.get()
    data_validator.check_saved_artifact_data(
        flow,
        table_artifact.id(),
        expected_data=pd.concat([extracted_table_data, extracted_table_data], ignore_index=True),
    )

    tables = flow.list_saved_objects()
    assert table_name in [item.spec.identifier() for item in tables[data_resource]]

    # Doesn't work if don't force because it is created in append mode.
    with pytest.raises(InvalidRequestError):
        client.delete_flow(flow.id(), saved_objects_to_delete=tables, force=False)

    # Actually delete the flow.
    client.delete_flow(flow.id(), saved_objects_to_delete=tables, force=True)

    # Check that the flow can not longer be accessed.
    with pytest.raises(InvalidUserArgumentException, match="Unable to find a flow"):
        client.flow(flow.id())

    # Check table no longer exists
    data_validator.check_saved_artifact_data_does_not_exist(table_artifact.id())


@pytest.mark.enable_only_for_data_resource_type(*all_relational_DBs())
# TODO(ENG-2881)
@pytest.mark.skip_for_spark_engines(reason="Need to investigate, TODO above.")
def test_delete_workflow_saved_objects_twice(
    client, flow_name, data_resource, engine, data_validator
):
    """Checking the successful deletion case and unsuccessful deletion case works as expected.
    To test this, I have two workflows that write to the same table. When I delete the table in the first workflow,
    it is successful but when I delete it in the second workflow, it is unsuccessful because the table has already
    been deleted.
    """
    table_name = generate_table_name()

    table_artifact = data_resource.sql("select * from hotel_reviews limit 5")
    save(data_resource, table_artifact, name=table_name, update_mode=LoadUpdateMode.REPLACE)

    # Workflow 1's name not specified, so given a random workflow name.
    flow1 = publish_flow_test(
        client,
        table_artifact,
        name=flow_name(),
        engine=engine,
    )

    # Workflow 2's name not specified, so given a random workflow name.
    save(data_resource, table_artifact, name=table_name, update_mode=LoadUpdateMode.APPEND)
    flow2 = publish_flow_test(
        client,
        table_artifact,
        name=flow_name(),
        engine=engine,
    )

    extracted_table_data = table_artifact.get()
    data_validator.check_saved_artifact_data(
        flow1,
        table_artifact.id(),
        expected_data=pd.concat([extracted_table_data, extracted_table_data], ignore_index=True),
    )

    tables = client.flow(flow1.id()).list_saved_objects()
    tables_1 = set([item.spec.identifier() for item in tables[data_resource]])
    assert table_name in tables_1

    tables = client.flow(flow2.id()).list_saved_objects()
    tables_2 = set([item.spec.identifier() for item in tables[data_resource]])
    assert table_name in tables_2

    assert tables_1 == tables_2
    client.delete_flow(flow1.id(), saved_objects_to_delete=tables, force=True)

    # Check flow indeed deleted and that the table no longer exists.
    with pytest.raises(InvalidUserArgumentException, match="Unable to find a flow"):
        client.flow(flow1.id())

    data_validator.check_saved_artifact_data_does_not_exist(table_artifact.id())

    # Try to delete table deleted by other flow.
    with pytest.raises(Exception) as e_info:
        client.delete_flow(flow2.id(), saved_objects_to_delete=tables, force=True)
    assert str(e_info.value).startswith("Failed to delete")

    # Failed to delete tables, but flow should be removed.
    with pytest.raises(InvalidUserArgumentException, match="Unable to find a flow"):
        client.flow(flow2.id())
