import uuid

from aqueduct.enums import ServiceType

from dataclasses import dataclass

@dataclass
class Table:
    id: uuid.UUID
    name: str
    service: ServiceType
    table: str

    def __init__(self, id: uuid.UUID, name: str, service: ServiceType, table: str):
        self.id = id
        self.name = name
        self.service = service
        self.table = table
    
    def to_dict(self) -> dict:
        return {
            "id": str(self.id),
            "name": self.name,
            "service": self.service.value,
            "table": self.table
        }
