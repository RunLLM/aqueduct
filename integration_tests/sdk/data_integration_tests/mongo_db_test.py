from typing import Optional

import pandas as pd
import pytest
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.constants.enums import ArtifactType
from aqueduct.error import AqueductError, InvalidUserArgumentException
from aqueduct.integrations.mongodb_integration import MongoDBIntegration

from sdk.data_integration_tests.flow_manager import FlowManager
from sdk.data_integration_tests.mongo_db_data_validator import MongoDBDataValidator
from sdk.data_integration_tests.save import save
from sdk.data_integration_tests.validation_helpers import (
    check_hotel_reviews_table_artifact,
    check_hotel_reviews_table_data,
)
from sdk.shared.globals import artifact_id_to_saved_identifier
from sdk.shared.naming import generate_object_name, generate_table_name
from sdk.shared.validation import check_artifact_was_computed

@pytest.fixture(autouse=True)
def assert_data_integration_is_mongo_db(data_integration):
    assert isinstance(data_integration, MongoDBIntegration)

def _save_artifact_and_check(
    flow_manager: FlowManager,
    data_integration: MongoDBIntegration,
    artifact: BaseArtifact,
    format: Optional[str],
    object_identifier: Optional[str] = None,
):
    """Saves the artifact by publishing a flow, and then checks that the data now exists in S3."""
    assert isinstance(artifact, BaseArtifact)

    if object_identifier is None:
        object_identifier = generate_table_name() if format is not None else generate_object_name()
    save(data_integration, artifact, object_identifier, format)

    flow = flow_manager.publish_flow_test(artifact)

    MongoDBDataValidator(flow_manager._client, data_integration).check_saved_artifact_data(
        flow, artifact.id(), artifact.type(), format, expected_data=artifact.get()
    )