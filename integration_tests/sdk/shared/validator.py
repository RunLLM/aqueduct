import uuid
from typing import Any, List, Tuple

import pandas as pd
from aqueduct.constants.enums import LoadUpdateMode
from aqueduct.integrations.sql_integration import RelationalDBIntegration
from aqueduct.models.integration import Integration
from aqueduct.models.operators import RelationalDBLoadParams

from aqueduct import Client, Flow

from .globals import artifact_id_to_saved_identifier
from .utils import extract


class Validator:
    """Tests can request an instance of this class as a fixture, and use it to validate published flow runs."""

    _client: Client
    _integration: Integration

    def __init__(self, client: Client, integration: Integration):
        self._client = client
        self._integration = integration

    def _fetch_saved_object_identifier(self, flow: Flow, artifact_id: uuid.UUID) -> str:
        """Also validates that the saved object exists according to the Flow API."""
        all_saved_objects = flow.list_saved_objects()[self._integration.name()]
        all_saved_object_identifiers = [item.spec.identifier() for item in all_saved_objects]

        saved_object_identifier = artifact_id_to_saved_identifier[str(artifact_id)]
        assert saved_object_identifier in all_saved_object_identifiers
        return saved_object_identifier

    def check_saved_artifact_data(
        self, flow: Flow, artifact_id: uuid.UUID, expected_data: Any
    ) -> None:
        """Checks that the given artifact was saved by the flow, and the data integration has the expected data.

        The exact destination of the artifact is tracked internally by the test suite.
        """
        assert expected_data is not None
        saved_object_identifier = self._fetch_saved_object_identifier(flow, artifact_id)

        # Verify the artifact's actual data state in the data integration.
        saved_data = extract(self._integration, saved_object_identifier).get()
        if not isinstance(saved_data, pd.DataFrame):
            raise Exception(
                "This test suite is expected to only deal with pandas Dataframe types."
                "For more extensive third-party type coverage, please write data integration "
                "tests instead."
            )
        assert isinstance(saved_data, pd.DataFrame)
        if not saved_data.equals(expected_data):
            print("Expected data: ", expected_data)
            print("Actual data: ", saved_data)
            raise Exception("Mismatch between expected and actual saved data.")

    def check_saved_artifact_data_does_not_exist(self, artifact_id: uuid.UUID) -> None:
        """Checks that the data integration no longer has the artifact's data stored."""
        saved_object_identifier = artifact_id_to_saved_identifier[str(artifact_id)]
        try:
            extract(self._integration, saved_object_identifier)
        except Exception:
            return

        raise Exception("Artifact %s is expected to no longer exist, but does." % artifact_id)

    def check_saved_update_mode_changes(
        self,
        flow: Flow,
        expected_updates: List[Tuple[str, LoadUpdateMode]],
        order_matters: bool = True,
    ):
        """Checks the exact result of flow.list_saved_objects().

        NOTE: This should only ever be called when checking saves against relational databases!

        When `order_matters=True`, the provided `expected_updates` list must match the fetched result exactly.
        Note that the updates are typically ordered from most to least recent.
        """
        assert isinstance(
            self._integration, RelationalDBIntegration
        ), "Currently, only relational data integrations are supported."

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

    def check_artifact_was_computed(self, flow: Flow, name: str):
        """Checks only for the artifact computed in the latest flow run."""
        artifact = flow.latest().artifact(name)
        assert artifact is not None
