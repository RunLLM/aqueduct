import uuid
from typing import Any, Dict, Optional

from aqueduct.responses import ArtifactResult
from aqueduct.dag import DAG, SubgraphDAGDelta, UpdateParametersDelta, apply_deltas_to_dag
from aqueduct import api_client
from aqueduct.generic_artifact import Artifact
from aqueduct.deserialize import deserialization_function_mapping
import aqueduct.table_artifact
import aqueduct.numeric_artifact
import aqueduct.bool_artifact
from aqueduct.enums import ArtifactType


def preview_artifact(dag: DAG, artifact_id: uuid.UUID, parameters: Optional[Dict[str, Any]] = None) -> Artifact:
    subgraph = apply_deltas_to_dag(
        dag,
        deltas=[
            SubgraphDAGDelta(
                artifact_ids=[artifact_id],
                include_load_operators=False,
            ),
            UpdateParametersDelta(
                parameters=parameters,
            ),
        ],
        make_copy=True,
    )

    preview_resp = api_client.__GLOBAL_API_CLIENT__.preview(dag=subgraph)
    artifact_response = preview_resp.artifact_results[artifact_id]

    serialization_type = artifact_response.serialization_type
    if serialization_type not in deserialization_function_mapping:
        raise Exception("Unsupported serialization type %s." % serialization_type)

    artifact_type = artifact_response.artifact_type
    content = deserialization_function_mapping[serialization_type](artifact_response.content)
    
    if artifact_type == ArtifactType.TABULAR:
        return aqueduct.table_artifact.TableArtifact(dag, artifact_id, content)
    elif artifact_type == ArtifactType.NUMERIC:
        return aqueduct.numeric_artifact.NumericArtifact(dag, artifact_id, content)
    elif artifact_type == ArtifactType.BOOL:
        return aqueduct.bool_artifact.BoolArtifact(dag, artifact_id, content)
    else:
        raise Exception("Unimplemented preview result artifact type!")
