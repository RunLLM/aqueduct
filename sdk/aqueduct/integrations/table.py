import uuid
from dataclasses import dataclass
from typing import Dict

from aqueduct.enums import ServiceType


@dataclass
class Table:
    table: str
    update_mode: str

    def __init__(self, table: str, update_mode: str):
        self.table = table
        self.update_mode = update_mode

    def to_dict(self) -> Dict[str, str]:
        return {"table": self.table, "update_mode": self.update_mode}
