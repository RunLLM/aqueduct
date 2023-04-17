import uuid
from typing import Any

import pandas as pd

from aqueduct import Client, Flow
from aqueduct.integrations.mongodb_integration import MongoDBIntegration
from sdk.data_integration_tests.relational_data_validator import RelationalDataValidator

from ..shared.validation import fetch_and_validate_saved_object_identifier


# MongoDBDataValidator inherits RelationalDataValidator to reuse `check_saved_update_mode_changes`
class MongoDBDataValidator(RelationalDataValidator):
    """Tests can request an instance of this class as a fixture, and use it to validate published flow runs."""

    _client: Client
    _integration: MongoDBIntegration

    def __init__(self, client: Client, integration: MongoDBIntegration):
        super(MongoDBDataValidator, self).__init__(client, integration)

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
        saved_data = self._integration.collection(saved_object_identifier).find({}).get()
        assert isinstance(saved_data, pd.DataFrame)
        if not saved_data.equals(expected_data):
            print("Expected data: ", expected_data)
            print("Actual data: ", saved_data)
            raise Exception("Mismatch between expected and actual saved data.")
