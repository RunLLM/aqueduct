import uuid
from typing import Any, Dict, Optional

from aqueduct.responses import ArtifactResult
from aqueduct.dag import DAG, SubgraphDAGDelta, UpdateParametersDelta, apply_deltas_to_dag
from aqueduct import api_client


def preview_artifact(dag: DAG, artifact_id: uuid.UUID, parameters: Optional[Dict[str, Any]] = None) -> ArtifactResult:
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
    return preview_resp.artifact_results[artifact_id]
