import base64
import json
import sys

from aqueduct_executor.migrators.parameter_val_type_inference_000019 import serialize
from aqueduct_executor.operators.utils.enums import SerializationType
from aqueduct_executor.operators.utils.utils import infer_artifact_type


def run_type_check_and_encode(json_val: str) -> None:
    """
    Infers the type for a given param value and print to std out
    Encode to base 64 and print out as well
    """
    val = json.loads(json_val)
    artifact_type = infer_artifact_type(val)
    serialization_type = serialize.artifact_type_to_serialization_type(artifact_type, val)
    print(serialization_type.value)
    print(serialize.serialize_val(val, serialization_type))


def run_decode(val: str, val_type: str) -> None:
    """
    Decodes from base64 and prints out the value
    """
    decoded_val = base64.b64decode(val)
    serialization_type = SerializationType(val_type)
    deserialized_val = serialize.deserialization_function_mapping[serialization_type](decoded_val)
    loaded_val = json.loads(json_val)
    print(json.dumps(loaded_val))
