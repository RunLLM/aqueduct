import json
import uuid

from pydantic import BaseModel

from aqueduct.constants.enums import ServiceType


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

    def is_relational(self) -> bool:
        """Returns whether the integration connects to a relational data store."""
        return self.service in [
            ServiceType.POSTGRES,
            ServiceType.SNOWFLAKE,
            ServiceType.MYSQL,
            ServiceType.REDSHIFT,
            ServiceType.MARIADB,
            ServiceType.SQLSERVER,
            ServiceType.BIGQUERY,
            ServiceType.AQUEDUCTDEMO,
            ServiceType.SQLITE,
            ServiceType.ATHENA,
        ]