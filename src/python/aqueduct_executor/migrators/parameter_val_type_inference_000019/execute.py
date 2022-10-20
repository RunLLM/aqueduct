import base64
import json

from aqueduct.utils import infer_artifact_type
from aqueduct_executor.migrators.parameter_val_type_inference_000019 import serialize
from aqueduct_executor.operators.utils.enums import SerializationType


def run_type_inference_and_encode(json_val: str) -> None:
    """
    Infers the type for a given param value and print to stdout
    Encode to base64 and print out as well
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
    print(json.dumps(deserialized_val))
