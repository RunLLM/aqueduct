import uuid
from typing import Any, Optional

import pandas as pd
from aqueduct.constants.enums import ArtifactType
from aqueduct.resources.s3 import S3Resource
from PIL import Image

from aqueduct import Client, Flow
from sdk.shared.validation import fetch_and_validate_saved_object_identifier


class S3DataValidator:
    _client: Client
    _resource: S3Resource

    def __init__(self, client: Client, resource: S3Resource):
        self._client = client
        self._resource = resource

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
            self._resource, flow, artifact_id
        )

        saved_artifact = self._resource.file(saved_object_identifier, artifact_type, format)
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
        elif isinstance(saved_data, list) and all(
            isinstance(elem, pd.DataFrame) for elem in saved_data
        ):
            is_equal = all(elem.equals(expected_data[i]) for i, elem in enumerate(saved_data))
        else:
            is_equal = saved_data == expected_data

        if not is_equal:
            print("Expected data: ", expected_data)
            print("Actual data: ", saved_data)
            raise Exception("Mismatch between expected and actual saved data.")
