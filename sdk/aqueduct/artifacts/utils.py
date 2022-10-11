from __future__ import annotations

import uuid
from typing import TYPE_CHECKING, Any, Dict, List, Optional, Union

from aqueduct.artifacts import bool_artifact, generic_artifact, numeric_artifact, table_artifact
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.dag import DAG
from aqueduct.dag_deltas import SubgraphDAGDelta, UpdateParametersDelta, apply_deltas_to_dag
from aqueduct.enums import ArtifactType
from aqueduct.error import InvalidArtifactTypeException
from aqueduct.responses import ArtifactResult
from aqueduct.serialization import deserialize
from aqueduct.utils import infer_artifact_type

from aqueduct import globals


def preview_artifact(
    dag: DAG, target_artifact_id: uuid.UUID, parameters: Optional[Dict[str, Any]] = None
) -> BaseArtifact:
    """Previews the given artifact and returns the resulting Artifact class.

    Will handle all type inference of the target artifact, as well as any upstream artifacts
    that were lazily computed.
    """
    return preview_artifacts(dag, [target_artifact_id], parameters)[0]


def preview_artifacts(
    dag: DAG, target_artifact_ids: List[uuid.UUID], parameters: Optional[Dict[str, Any]] = None
) -> List[BaseArtifact]:
    """Batch version of `preview_artifact()`

    Returns a list of artifacts, each corresponding to one of the provided `target_artifact_ids`, in
    the same order.
    """
    subgraph = apply_deltas_to_dag(
        dag,
        deltas=[
            SubgraphDAGDelta(
                artifact_ids=target_artifact_ids,
                include_saves=False,
            ),
            UpdateParametersDelta(
                parameters=parameters,
            ),
        ],
        make_copy=True,
    )

    preview_resp = globals.__GLOBAL_API_CLIENT__.preview(dag=subgraph)

    # Process all the target artifacts first. Assumption: the preview response contains a result entry
    # for each target artifact.
    output_artifacts: List[BaseArtifact] = []
    for target_artifact_id in target_artifact_ids:
        assert (
            target_artifact_id in preview_resp.artifact_results.keys()
        ), "Preview is expected to return a result for each target artifact."
        target_artifact_result = preview_resp.artifact_results[target_artifact_id]

        # Fetch the inferred type of the target artifact.
        target_artifact_type = target_artifact_result.artifact_type
        _update_artifact_type(dag, target_artifact_id, target_artifact_type)

        # Fetch the content of the target artifact.
        artifact_name = dag.must_get_artifact(target_artifact_id).name
        target_artifact_content = _get_content_from_artifact_result_resp(
            target_artifact_result, artifact_name
        )

        # Create the target artifact.
        output_artifacts.append(
            to_artifact_class(
                dag,
                target_artifact_id,
                target_artifact_type,
                target_artifact_content,
            )
        )

    # Any non-target artifacts are guaranteed to be upstream of the target artifact (due to the SubgraphDAGDelta),
    # so if any of them are the result of a lazy operation, we'll want to backfill their types. *We do NOT backfill
    # their contents*, as those are stored on the Artifact class itself and not the underlying shared dag.
    for artifact_id, artifact_result in preview_resp.artifact_results.items():
        # We've already processed the target artifact.
        if artifact_id in target_artifact_ids:
            continue

        _update_artifact_type(dag, artifact_id, artifact_result.artifact_type)

    return output_artifacts


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


def _get_content_from_artifact_result_resp(
    artifact_result: ArtifactResult, artifact_name: str
) -> Any:
    """Deserialize and validate the type of the content for a given artifact result."""
    content = deserialize(
        artifact_result.serialization_type,
        artifact_result.artifact_type,
        artifact_result.content,
    )
    assert infer_artifact_type(content) == artifact_result.artifact_type, (
        "Artifact `%s` has deserialized content with type %s, but preview request returned type %s"
        % (
            artifact_name,
            infer_artifact_type(content),
            artifact_result.artifact_type,
        )
    )
    return content
