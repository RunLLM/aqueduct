import uuid

from aqueduct.models.resource import BaseResource

from aqueduct import Client, Flow
from sdk.shared.globals import artifact_id_to_saved_identifier

"""These helpers are shared across both integration test suites."""


def fetch_and_validate_saved_object_identifier(
    data_integration: BaseResource, flow: Flow, artifact_id: uuid.UUID
) -> str:
    """Validates that the saved object exists according to the Flow API."""
    all_saved_objects = flow.list_saved_objects()[data_integration.name()]
    all_saved_object_identifiers = [item.spec.identifier() for item in all_saved_objects]

    saved_object_identifier = artifact_id_to_saved_identifier[str(artifact_id)]
    assert saved_object_identifier in all_saved_object_identifiers
    return saved_object_identifier


def check_artifact_was_computed(flow: Flow, name: str):
    """Checks only for the artifact computed in the latest flow run."""
    artifact = flow.latest().artifact(name)
    assert artifact is not None
