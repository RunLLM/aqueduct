import uuid
from typing import Any

import pandas as pd
from aqueduct.models.integration import BaseResource

from aqueduct import Client, Flow

from ..shared.globals import artifact_id_to_saved_identifier
from ..shared.validation import fetch_and_validate_saved_object_identifier
from .extract import extract


class DataValidator:
    """Tests can request an instance of this class as a fixture, and use it to validate published flow runs."""

    _client: Client
    _integration: BaseResource

    def __init__(self, client: Client, integration: BaseResource):
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
        saved_data = extract(self._integration, saved_object_identifier).get()
        if not isinstance(saved_data, pd.DataFrame):
            raise Exception(
                "This method is expected to only deal with pandas Dataframe types."
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
