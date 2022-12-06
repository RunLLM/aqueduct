import uuid
from typing import Any, List, Tuple

from aqueduct import Flow, LoadUpdateMode, Client
from aqueduct.integrations.integration import Integration
from aqueduct.integrations.sql_integration import RelationalDBIntegration
from utils import artifact_id_to_saved_identifier


class Validator():
    """TODO: describe what these validators are trying to do."""
    _client: Client
    _integration: Integration

    def __init__(self, client: Client, integration: Integration):
        self._client = client
        self._integration = integration

    def check_saved_artifact(self, flow: Flow, artifact_id: uuid.UUID, expected_data: Any):
        """TODO: talk about limitations across workflow runs.

        # TODO: this only works with SQL-based integrations!
        """
        assert isinstance(self._integration, RelationalDBIntegration), "Currently, only relational data integrations are supported."

        # Check that given saved artifacts were indeed saved based on the flow API.
        all_saved_objects = flow.list_saved_objects()[self._integration._metadata.name]
        all_saved_object_identifiers = [item.object_name for item in all_saved_objects]
        saved_object_identifier = artifact_id_to_saved_identifier[str(artifact_id)]
        assert saved_object_identifier in all_saved_object_identifiers

        # Verify that the actual data was saved.
        saved_data = self._integration.sql(f"SELECT * FROM {saved_object_identifier}").get()
        if not saved_data.equals(expected_data):
            print("Expected data: ", expected_data)
            print("Actual data: ", saved_data)
            raise Exception("Mismatch between expected and actual saved data.")

    def check_saved_update_mode_changes(self, flow: Flow, expected_updates: List[Tuple[str, LoadUpdateMode]], order_matters: bool = True):
        # TODO: table_name, update_mode in order of latest created

        data = self._client.flow(flow.id()).list_saved_objects()

        # Check all objects were saved to the same integration.
        assert len(data.keys()) == 1
        integration_name = list(data.keys())[0]
        assert integration_name == self._integration._metadata.name

        assert len(data[integration_name]) == len(expected_updates)
        saved_objects = data[integration_name]
        actual_updates = [(saved_objects[i].object_name, saved_objects[i].update_mode) for i, (name, _) in enumerate(expected_updates)]

        if order_matters:
            assert expected_updates == actual_updates, "Expected %s, got %s." % (expected_updates, actual_updates)
        else:
            assert all(actual_update in expected_updates for actual_update in actual_updates)

        # Check that mapping can be accessed by integration object too.
        assert data[self._integration] == data[integration_name]