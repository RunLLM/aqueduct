from __future__ import annotations

import uuid
from typing import TYPE_CHECKING, Any, Dict, Optional, Union

from aqueduct.artifacts import bool_artifact, generic_artifact, numeric_artifact, table_artifact
from aqueduct.dag import DAG
from aqueduct.dag_deltas import (
    SubgraphDAGDelta,
    UpdateParametersDelta,
    apply_deltas_to_dag,
    validate_overwriting_parameters,
)
from aqueduct.enums import ArtifactType
from aqueduct.error import InvalidArtifactTypeException
from aqueduct.responses import ArtifactResult
from aqueduct.serialization import deserialization_function_mapping
from aqueduct.utils import infer_artifact_type

from aqueduct import globals

if TYPE_CHECKING:
    from aqueduct.artifacts.bool_artifact import BoolArtifact
    from aqueduct.artifacts.generic_artifact import GenericArtifact
    from aqueduct.artifacts.numeric_artifact import NumericArtifact
    from aqueduct.artifacts.table_artifact import TableArtifact


def preview_artifact(
    dag: DAG, target_artifact_id: uuid.UUID, parameters: Optional[Dict[str, Any]] = None
) -> Union[TableArtifact, NumericArtifact, BoolArtifact, GenericArtifact]:
    """Previews the given artifact and returns the resulting Artifact class.

    Will handle all type inference of the target artifact, as well as any upstream artifacts
    that were lazily computed.
    """
    subgraph = apply_deltas_to_dag(
        dag,
        deltas=[
            SubgraphDAGDelta(
                artifact_ids=[target_artifact_id],
                include_load_operators=False,
            ),
            UpdateParametersDelta(
                parameters=parameters,
            ),
        ],
        make_copy=True,
    )

    preview_resp = globals.__GLOBAL_API_CLIENT__.preview(dag=subgraph)

    assert target_artifact_id in preview_resp.artifact_results.keys()
    target_artifact_result = preview_resp.artifact_results[target_artifact_id]
    target_artifact_content = _get_content_from_artifact_result_resp(target_artifact_result)
    target_artifact_type = target_artifact_result.artifact_type

    _update_artifact_type(dag, target_artifact_id, target_artifact_type)

    # Any non-target artifacts are guaranteed to be upstream of the target artifact (due to the SubgraphDAGDelta),
    # so if any of them are the result of a lazy operation, we'll want to backfill their types. *We do NOT backfill
    # their contents*, as those are stored on the Artifact class itself and not the underlying shared dag.
    for artifact_id, artifact_result in preview_resp.artifact_results.items():
        # We've already processed the target artifact.
        if artifact_id == target_artifact_id:
            continue

        _update_artifact_type(dag, artifact_id, artifact_result.artifact_type)

    return to_artifact_class(dag, target_artifact_id, target_artifact_type, target_artifact_content)


def to_artifact_class(
    dag: DAG,
    artifact_id: uuid.UUID,
    artifact_type: ArtifactType = ArtifactType.UNTYPED,
    content: Optional[Any] = None,
) -> Union[TableArtifact, NumericArtifact, BoolArtifact, GenericArtifact]:
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


def _update_artifact_type(
    dag: DAG, artifact_id: uuid.UUID, new_artifact_type: ArtifactType
) -> None:
    """Update's the type for an untyped artifact in the DAG.

    Fails if there is a type mismatch with an already existing type. This is safe to use as much
    as you want (eg. for backfills), since it will not change types arbitrarily.
    """
    artifact = dag.must_get_artifact(artifact_id)
    existing_type_annotation = artifact.type
    if (
        existing_type_annotation != ArtifactType.UNTYPED
        and existing_type_annotation != new_artifact_type
    ):
        raise InvalidArtifactTypeException(
            "The artifact `%s` was expected to have type %s, but instead computed type %s"
            % (artifact.name, existing_type_annotation, new_artifact_type)
        )

    # If the artifact was already typed, we've verified that the type does not change above.
    if existing_type_annotation == ArtifactType.UNTYPED:
        dag.update_artifact_type(artifact_id, new_artifact_type)


def _get_content_from_artifact_result_resp(artifact_result: ArtifactResult) -> Any:
    """Deserialize and validate the type of the content for a given artifact result."""
    serialization_type = artifact_result.serialization_type
    if serialization_type not in deserialization_function_mapping:
        raise Exception("Unsupported serialization type %s." % serialization_type)

    content = deserialization_function_mapping[serialization_type](artifact_result.content)
    assert infer_artifact_type(content) == artifact_result.artifact_type
    return content
