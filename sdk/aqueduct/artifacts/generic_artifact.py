from __future__ import annotations

import uuid
from typing import Any, Dict, Optional

from aqueduct.artifacts import utils as artifact_utils
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.dag import DAG
from aqueduct.enums import ArtifactType
from aqueduct.error import AqueductError


class GenericArtifact(BaseArtifact):
    """This class represents a generic artifact within the flow's DAG.
    Currently, a generic artifact can be any artifact other than table, numeric, bool, or parameter.
    """

    def __init__(
        self,
        dag: DAG,
        artifact_id: uuid.UUID,
        type: ArtifactType,
        content: Optional[Any] = None,
        from_flow_run: bool = False,
    ):
        self._dag = dag
        self._artifact_id = artifact_id
        # This parameter indicates whether the artifact is fetched from flow-run or not.
        self._from_flow_run = from_flow_run
        self._content = content
        if self._from_flow_run:
            # If the artifact is initialized from a flow run, then it should not contain any content.
            assert self._content is None
        else:
            assert self._content is not None

        self._type = type

    def get(self, parameters: Optional[Dict[str, Any]] = None) -> Any:
        """Materializes the artifact.

        Returns:
            The materialized value.

        Raises:
            InvalidRequestError:
                An error occurred because of an issue with the user's code or inputs.
            InternalServerError:
                An unexpected error occurred in the server.
        """
        self._dag.must_get_artifact(self._artifact_id)

        if parameters:
            artifact = artifact_utils.preview_artifact(self._dag, self._artifact_id, parameters)
            if artifact.type() != ArtifactType.BOOL:
                raise Exception(
                    "Error: the computed result is expected to of type bool, found %s"
                    % artifact.type()
                )
            return artifact._content

        if self._content is None:
            self._content = artifact_utils.preview_artifact(self._dag, self._artifact_id)._content

        return self._content

    def describe(self) -> None:
        """Prints out a human-readable description of the bool artifact."""
        # TODO: make this more informative.
        print("This is a %s artifact." % self.type())
