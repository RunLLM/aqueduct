import uuid
from abc import ABC
import json

from pydantic import BaseModel

from aqueduct.enums import ServiceType


class IntegrationInfo(BaseModel):
    id: uuid.UUID
    name: str
    service: ServiceType
    createdAt: int
    validated: bool

    def describe(self) -> None:
        """Prints out a human-readable description of the integration."""
        description_map = {
            "Id": str(self.id),
            "Name": self.name,
            "Service": self.service,
            "CreatedAt": self.createdAt,
            "Validated": self.validated,
        }
        print(json.dumps(description_map, sort_keys=False, indent=4))


class Integration(ABC):
    """
    Abstract class for the various integrations Aqueduct interacts with.
    """

    _metadata: IntegrationInfo
