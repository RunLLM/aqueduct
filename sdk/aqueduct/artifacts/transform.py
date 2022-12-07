from __future__ import annotations

import uuid
from typing import Any, Optional

from aqueduct.artifacts import bool_artifact, generic_artifact, numeric_artifact, table_artifact
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.constants.enums import ArtifactType
from aqueduct.models.dag import DAG


def to_artifact_class(
    dag: DAG,
    artifact_id: uuid.UUID,
    artifact_type: ArtifactType = ArtifactType.UNTYPED,
    content: Optional[Any] = None,
) -> BaseArtifact:
    if artifact_type == ArtifactType.TABLE:
        return table_artifact.TableArtifact(
            dag,
            artifact_id,
            content,
        )
    elif artifact_type == ArtifactType.NUMERIC:
        return numeric_artifact.NumericArtifact(dag, artifact_id, content)
    elif artifact_type == ArtifactType.BOOL:
        return bool_artifact.BoolArtifact(dag, artifact_id, content)
    else:
        return generic_artifact.GenericArtifact(dag, artifact_id, artifact_type, content)
