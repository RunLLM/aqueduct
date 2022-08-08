from __future__ import annotations

import uuid
from typing import Any, Dict, Optional

import cloudpickle as pickle
from aqueduct.api_client import APIClient
from aqueduct.dag import DAG, SubgraphDAGDelta, UpdateParametersDelta, apply_deltas_to_dag
from aqueduct.deserialize import deserialization_function_mapping
from aqueduct.enums import SerializationType
from aqueduct.error import AqueductError
from aqueduct.generic_artifact import Artifact
from aqueduct.operators import SaveConfig
from aqueduct.utils import format_header_for_print, get_description_for_check

from aqueduct import api_client


class UntypedArtifact(Artifact):
    """This class represents an artifact with unknown type within the flow's DAG."""

    def __init__(self, dag: DAG, artifact_id: uuid.UUID, from_flow_run: bool = False):
        self._dag = dag
        self._artifact_id = artifact_id
        # This parameter indicates whether the artifact is fetched from flow-run or not.
        self._from_flow_run = from_flow_run

    def get(self, parameters: Optional[Dict[str, Any]] = None) -> Any:
        """Materializes the untyped artifact.

        Returns:
            The materialized value of this artifact.

        Raises:
            InvalidRequestError:
                An error occurred because of an issue with the user's code or inputs.
            InternalServerError:
                An unexpected error occurred in the server.
        """
        dag = apply_deltas_to_dag(
            self._dag,
            deltas=[
                SubgraphDAGDelta(
                    artifact_ids=[self._artifact_id],
                    include_load_operators=False,
                ),
                UpdateParametersDelta(
                    parameters=parameters,
                ),
            ],
            make_copy=True,
        )
        preview_resp = api_client.__GLOBAL_API_CLIENT__.preview(dag=dag)
        artifact_response = preview_resp.artifact_results[self._artifact_id]

        serialization_type = artifact_response.serialization_type

        if serialization_type not in deserialization_function_mapping:
            raise Exception("Unsupported serialization type %s." % serialization_type)

        return deserialization_function_mapping[serialization_type](artifact_response.content)

    def describe(self) -> None:
        """Prints out a human-readable description of the check artifact."""
        print("Describe method not implemented yet for untyped artifact.")

    def save(self, config: SaveConfig) -> None:
        print("Save method not implemented yet for untyped artifact.")
