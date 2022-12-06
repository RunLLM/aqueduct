from abc import ABC
from typing import Any

from aqueduct.models.integration_info import IntegrationInfo


class Integration(ABC):
    """
    Abstract class for the various integrations Aqueduct interacts with.
    """

    _metadata: IntegrationInfo

    def __hash__(self) -> int:
        """An integration is uniquely identified by its name.
        Ref: https://docs.python.org/3.5/reference/datamodel.html#object.__hash__
        """
        return hash(self._metadata.name)

    def __eq__(self, other: Any) -> bool:
        """The string and Integration object representation are equivalent allowing
        the user to access a dictionary keyed by the Integration object with the
        integration name as a string and vice versa
        """
        if type(other) == type(self) and "name" in other._metadata.__dict__:
            return bool(self._metadata.name == other._metadata.name)
        elif type(other) == str:
            return bool(self._metadata.name == other)
        return False
