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
            "Id": str(self.id),
            "Name": self.name,
            "Service": self.service.value,
            "Table": self.table
        }
