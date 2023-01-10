import pytest
from aqueduct.error import InvalidIntegrationException

from ..shared.data_objects import DataObject
from ..shared.utils import extract
from .save import save
from .test_functions.simple.model import dummy_sentiment_model


def test_invalid_source_integration(client):
    with pytest.raises(InvalidIntegrationException):
        client.integration(name="wrong integration name")


def test_invalid_destination_integration(client, data_integration):
    table_artifact = extract(data_integration, DataObject.SENTIMENT)
    output_artifact = dummy_sentiment_model(table_artifact)

    with pytest.raises(InvalidIntegrationException):
        data_integration._metadata.name = "bad name"
        save(data_integration, output_artifact)
