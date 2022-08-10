import base64
import uuid

from aqueduct.responses import ArtifactResult
from aqueduct.generic_artifact import Artifact
from aqueduct.enums import ArtifactType
from aqueduct.deserialize import deserialization_function_mapping
from aqueduct.table_artifact import TableArtifact
from aqueduct.numeric_artifact import NumericArtifact
from aqueduct.dag import DAG


def to_typed_artifact(dag: DAG, artifact_id: uuid.UUID, artifact_response: ArtifactResult) -> Artifact:
    serialization_type = artifact_response.serialization_type
    if serialization_type not in deserialization_function_mapping:
        raise Exception("Unsupported serialization type %s." % serialization_type)

    artifact_type = artifact_response.artifact_type
    content = deserialization_function_mapping[serialization_type](base64.b64decode(artifact_response.content))
    
    if artifact_type == ArtifactType.TABULAR:
        return TableArtifact(dag, artifact_id, content)
    elif artifact_type == ArtifactType.NUMERIC:
        return NumericArtifact(dag, artifact_id, content)
    else:
        raise Exception("Unimplemented preview result artifact type!")
