import base64
import json

from aqueduct_executor.migrators.backfill_python_type_000022 import serialize
from aqueduct_executor.migrators.backfill_python_type_000022.spec import MigrationSpec
from aqueduct_executor.operators.utils.storage.parse import parse_storage


def run(spec: MigrationSpec) -> None:
    storage = parse_storage(spec.storage_config)
    content = storage.get(spec.content_path)

    deserialized_content = serialize.deserialization_function_mapping[spec.serialization_type](
        content
    )
    print(type(deserialized_content).__name__)
