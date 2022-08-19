from __future__ import annotations

import base64
import uuid
from typing import TYPE_CHECKING, Any, Dict, Optional, Union

from aqueduct.artifacts import bool_artifact, generic_artifact, numeric_artifact, table_artifact
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.dag import DAG, SubgraphDAGDelta, UpdateParametersDelta, apply_deltas_to_dag
from aqueduct.deserialize import deserialization_function_mapping
from aqueduct.enums import ArtifactType
from aqueduct.responses import ArtifactResult

from aqueduct import api_client

if TYPE_CHECKING:
    from aqueduct.artifacts.bool_artifact import BoolArtifact
    from aqueduct.artifacts.generic_artifact import GenericArtifact
    from aqueduct.artifacts.numeric_artifact import NumericArtifact
    from aqueduct.artifacts.table_artifact import TableArtifact


def preview_artifact(
    dag: DAG, artifact_id: uuid.UUID, parameters: Optional[Dict[str, Any]] = None
) -> Union[TableArtifact, NumericArtifact, BoolArtifact, GenericArtifact,]:
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
    content = deserialization_function_mapping[serialization_type](
        base64.b64decode(artifact_response.content)
    )

    if artifact_type == ArtifactType.TABLE:
        return table_artifact.TableArtifact(dag, artifact_id, content)
    elif artifact_type == ArtifactType.NUMERIC:
        return numeric_artifact.NumericArtifact(dag, artifact_id, content)
    elif artifact_type == ArtifactType.BOOL:
        return bool_artifact.BoolArtifact(dag, artifact_id, content)
    else:
        return generic_artifact.GenericArtifact(dag, artifact_id, artifact_type, content)
