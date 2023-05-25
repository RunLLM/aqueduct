from __future__ import annotations

import uuid
from typing import Any, Dict, List, Optional

from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.create import to_artifact_class
from aqueduct.constants.enums import ArtifactType, K8sClusterStatusType
from aqueduct.error import InvalidArtifactTypeException
from aqueduct.models.config import EngineConfig
from aqueduct.models.dag import DAG
from aqueduct.models.response_models import ArtifactResult
from aqueduct.utils.dag_deltas import SubgraphDAGDelta, UpdateParametersDelta, apply_deltas_to_dag
from aqueduct.utils.serialization import deserialize
from aqueduct.utils.type_inference import infer_artifact_type
from aqueduct.utils.utils import generate_engine_config

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
    global_engine_config: Optional[EngineConfig] = None
    if globals.__GLOBAL_CONFIG__.engine is not None:
        global_engine_config = generate_engine_config(
            globals.__GLOBAL_API_CLIENT__.list_resources(),
            globals.__GLOBAL_CONFIG__.engine,
        )

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
    subgraph.set_engine_config(global_engine_config=global_engine_config)

    engine_statuses = globals.__GLOBAL_API_CLIENT__.get_dynamic_engine_status_by_dag(dag=subgraph)
    for name in engine_statuses:
        engine_status = engine_statuses[name]
        if engine_status.status != K8sClusterStatusType.ACTIVE:
            print(
                "Your preview request makes use of dynamic k8s resource %s, but the k8s cluster is in %s state. It could take 12-15 minutes for the cluster to be ready..."
                % (engine_status.name, engine_status.status.value)
            )
        else:
            print(
                "Your preview request makes use of dynamic k8s resource %s and the k8s cluster is in %s state, so you are good to go!"
                % (engine_status.name, engine_status.status.value)
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
