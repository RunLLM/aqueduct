import uuid
from typing import Any, List, Tuple

import pandas as pd
from aqueduct.constants.enums import LoadUpdateMode
from aqueduct.integrations.sql_integration import RelationalDBIntegration
from aqueduct.models.operators import RelationalDBLoadParams

from aqueduct import Client, Flow

from ..shared.validation import fetch_and_validate_saved_object_identifier


class RelationalDataValidator:
    """Tests can request an instance of this class as a fixture, and use it to validate published flow runs."""

    _client: Client
    _integration: RelationalDBIntegration

    def __init__(self, client: Client, integration: RelationalDBIntegration):
        self._client = client
        self._integration = integration

    def check_saved_artifact_data(
        self, flow: Flow, artifact_id: uuid.UUID, expected_data: Any
    ) -> None:
        """Checks that the given artifact was saved by the flow, and the data integration has the expected data.

        The exact destination of the artifact is tracked internally by the test suite.
        """
        assert expected_data is not None
        saved_object_identifier = fetch_and_validate_saved_object_identifier(
            self._integration, flow, artifact_id
        )

        # Verify the artifact's actual data state in the data integration.
        saved_data = self._integration.sql(query="SELECT * from %s" % saved_object_identifier).get()
        assert isinstance(saved_data, pd.DataFrame)
        if not saved_data.equals(expected_data):
            print("Expected data: ", expected_data)
            print("Actual data: ", saved_data)
            raise Exception("Mismatch between expected and actual saved data.")

    def check_saved_update_mode_changes(
        self,
        flow: Flow,
        expected_updates: List[Tuple[str, LoadUpdateMode]],
        order_matters: bool = True,
    ):
        """Checks the exact result of flow.list_saved_objects().

        When `order_matters=True`, the provided `expected_updates` list must match the fetched result exactly.
        Note that the updates are typically ordered from most to least recent.
        """
        data = self._client.flow(flow.id()).list_saved_objects()

        # Check all objects were saved to the same integration.
        assert len(data.keys()) == 1
        integration_name = list(data.keys())[0]
        assert integration_name == self._integration.name()

        assert len(data[integration_name]) == len(expected_updates)
        saved_objects = data[integration_name]

        assert all(
            isinstance(saved_object.spec.parameters, RelationalDBLoadParams)
            for saved_object in saved_objects
        )
        actual_updates = [
            (saved_objects[i].spec.parameters.table, saved_objects[i].spec.parameters.update_mode)
            for i, (name, _) in enumerate(expected_updates)
        ]

        if order_matters:
            assert expected_updates == actual_updates, "Expected %s, got %s." % (
                expected_updates,
                actual_updates,
            )
        else:
            assert all(actual_update in expected_updates for actual_update in actual_updates)

        # Check that mapping can be accessed by integration object too.
        assert data[self._integration] == data[integration_name]
