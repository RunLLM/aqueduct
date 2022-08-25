from aqueduct_executor.operators.utils.storage.config import StorageConfig
from aqueduct_executor.operators.utils.storage.file import FileStorage
from aqueduct_executor.operators.utils.storage.gcs import GCSStorage
from aqueduct_executor.operators.utils.storage.s3 import S3Storage
from aqueduct_executor.operators.utils.storage.storage import Storage


def parse_storage(storage_config: StorageConfig) -> Storage:
    if storage_config.s3_config:
        return S3Storage(storage_config.s3_config)
    if storage_config.file_config:
        return FileStorage(storage_config.file_config)
    if storage_config.gcs_config:
        return GCSStorage(storage_config.gcs_config)
    raise Exception("Unknown storage type")
