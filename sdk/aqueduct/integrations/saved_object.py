import uuid
from dataclasses import dataclass
from typing import Dict

from aqueduct.enums import ServiceType


@dataclass
class SavedObject:
    name: str
    update_mode: str

    def __init__(self, name: str, update_mode: str):
        self.name = name
        self.update_mode = update_mode
