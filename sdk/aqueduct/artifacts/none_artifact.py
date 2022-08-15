from __future__ import annotations

import uuid
from typing import Any, Dict, Optional

from aqueduct.artifacts import utils as artifact_utils
from aqueduct.artifacts.artifact import Artifact
from aqueduct.dag import DAG
from aqueduct.enums import ArtifactType
from aqueduct.error import AqueductError


class NoneArtifact(Artifact):
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
        return None

    def describe(self) -> None:
        """Prints out a human-readable description of the bool artifact."""
        # TODO: make this more informative.
        print("This is a %s artifact." % self.type())
