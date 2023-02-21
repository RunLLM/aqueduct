import uuid
from typing import Any, Optional

import pandas as pd
from aqueduct.constants.enums import ArtifactType
from aqueduct.integrations.s3_integration import S3Integration
from PIL import Image

from aqueduct import Client, Flow
from sdk.shared.validation import fetch_and_validate_saved_object_identifier


class S3DataValidator:
    _client: Client
    _integration: S3Integration

    def __init__(self, client: Client, integration: S3Integration):
        self._client = client
        self._integration = integration

    def check_saved_artifact_data(
        self,
        flow: Flow,
        artifact_id: uuid.UUID,
        artifact_type: ArtifactType,
        format: Optional[str],
        expected_data: Any,
        skip_data_check: bool,
    ) -> None:
        assert expected_data is not None

        saved_object_identifier = fetch_and_validate_saved_object_identifier(
            self._integration, flow, artifact_id
        )

        saved_artifact = self._integration.file(saved_object_identifier, artifact_type, format)
        assert saved_artifact.type() == artifact_type
        saved_data = saved_artifact.get()
        assert type(saved_data) == type(expected_data), "Expected data type %s, get type %s." % (
            type(expected_data),
            type(saved_data),
        )

        if skip_data_check:
            return

        if isinstance(saved_data, pd.DataFrame):
            is_equal = saved_data.equals(expected_data)
        else:
            is_equal = saved_data == expected_data

        if not is_equal:
            print("Expected data: ", expected_data)
            print("Actual data: ", saved_data)
            raise Exception("Mismatch between expected and actual saved data.")
