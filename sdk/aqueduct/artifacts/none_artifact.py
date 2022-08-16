from __future__ import annotations

import uuid
from typing import Any, Dict, Optional

from aqueduct.artifacts.artifact import Artifact
from aqueduct.dag import DAG
from aqueduct.enums import ArtifactType


class NoneArtifact(Artifact):
    """This class represents a none artifact within the flow's DAG.
    Currently, a none artifact will only be created when there is no return value for the operator function
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
        """
        Returns:
            None
        """
        return None

    def describe(self) -> None:
        """Prints out a human-readable description of the none artifact."""
        # TODO: make this more informative.
        print("This is a %s artifact." % self.type())
