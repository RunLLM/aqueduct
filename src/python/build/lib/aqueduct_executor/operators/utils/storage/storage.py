from abc import ABC, abstractmethod


class Storage(ABC):
    @abstractmethod
    def put(self, key: str, value: bytes) -> None:
        pass

    @abstractmethod
    def get(self, key: str) -> bytes:
        pass
