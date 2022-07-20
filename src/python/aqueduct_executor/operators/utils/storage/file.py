from aqueduct_executor.operators.utils.storage.config import FileStorageConfig
from aqueduct_executor.operators.utils.storage.storage import Storage


class FileStorage(Storage):
    _config: FileStorageConfig

    def __init__(self, config: FileStorageConfig):
        self._config = config

    def put(self, key: str, value: bytes) -> None:
        path = self.get_full_path(key)
        print(f"writing to file: {path}")
        with open(path, "wb") as f:
            f.write(value)

    def get(self, key: str) -> bytes:
        path = self.get_full_path(key)
        print(f"reading from file: {path}")
        with open(path, "rb") as f:
            return f.read()

    def get_full_path(self, key: str) -> str:
        return self._config.directory + "/" + key
