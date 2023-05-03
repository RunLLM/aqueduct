import pytest
from aqueduct.constants.enums import ServiceType
from aqueduct.error import (
    AqueductError,
    InvalidIntegrationException,
    InvalidRequestError,
    InvalidUserActionException,
    InvalidUserArgumentException,
)
from aqueduct.resources.connect_config import K8sConfig
from pydantic import ValidationError

from aqueduct import global_config

from ..shared.data_objects import DataObject
from .extract import extract
from .save import save
from .test_functions.simple.model import dummy_sentiment_model


def test_invalid_source_integration(client):
    with pytest.raises(InvalidIntegrationException):
        client.resource(name="wrong integration name")


def test_invalid_destination_integration(data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)

    with pytest.raises(InvalidIntegrationException):
        data_integration._metadata.name = "bad name"
        save(data_integration, output_artifact)


def test_invalid_connect_integration(client):
    # Name already exists.
    config = {
        "database": "test",
    }
    with pytest.raises(
        InvalidUserActionException, match="An integration with this name already exists."
    ):
        client.connect_resource("aqueduct_demo", "SQLite", config)

    # Service is invalid.
    with pytest.raises(
        InvalidUserArgumentException,
        match="Service argument must match exactly one of the enum values in ServiceType.",
    ):
        client.connect_resource("New Resource", "invalid service", config)

    # Invalid config raises a pydantic error.
    with pytest.raises(ValidationError):
        client.connect_resource("New Resource", "SQLite", {})


@pytest.mark.enable_only_for_engine_type(ServiceType.K8S)
@pytest.mark.enable_only_for_data_integration_type(ServiceType.SQLITE)
def test_sqlite_with_k8s(data_integration, engine):
    """Tests that running an extract operator that reads data from a SQLite database using k8s should fail."""
    global_config({"engine": engine})
    with pytest.raises(AqueductError, match="Unknown integration service provided SQLite"):
        extract(data_integration, DataObject.SENTIMENT)


@pytest.mark.enable_only_for_local_storage()
def test_compute_integration_without_cloud_storage(client):
    with pytest.raises(
        InvalidRequestError,
        match="You need to setup cloud storage as metadata store before registering compute integration of type Kubernetes.",
    ):
        client.connect_resource(
            name="compute integration without cloud storage",
            service=ServiceType.K8S,
            config=K8sConfig(kubeconfig_path="dummy_path", cluster_name="dummy_name"),
        )


def test_cannot_delete_artifact_store_integration(client, artifact_store):
    # Skip test for local artifact storage.
    if artifact_store is None:
        return

    with pytest.raises(
        InvalidRequestError,
        match="Cannot delete an integration that is being used as artifact storage.",
    ):
        client.delete_resource(artifact_store)


# TODO (ENG-2593): Investigate ways to support relative kubeconfig and aws credential path
# def test_k8s_integration_wrong_kubeconfig(client):
#    with pytest.raises(InvalidRequestError):
#        client.connect_integration(
#            name="k8s integration with wrong kubeconfig",
#            service=ServiceType.K8S,
#            config=K8sConfig(kubeconfig_path="compute/k8s/wrong_kubeconfig", cluster_name="dummy_name"),
#        )
