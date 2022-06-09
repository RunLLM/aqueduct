from aqueduct_executor.operators.utils.storage.config import FileStorageConfig
from aqueduct_executor.operators.utils.storage.storage import Storage


class FileStorage(Storage):
    _config: FileStorageConfig

    def __init__(self, config: FileStorageConfig):
        self._config = config

    def put(self, key: str, value: bytes) -> None:
        print(f"writing to file: {key}")
        with open(self.get_full_path(key), "wb") as f:
            f.write(value)

    def get(self, key: str) -> bytes:
        print(f"reading from file: {key}")
        with open(self.get_full_path(key), "rb") as f:
            return f.read()

    def get_full_path(self, key: str) -> str:
        return self._config.directory + "/" + key
