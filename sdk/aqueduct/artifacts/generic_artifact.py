from __future__ import annotations

import uuid
from typing import Any, Dict, Optional

from aqueduct.artifacts import utils as artifact_utils
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.dag import DAG
from aqueduct.enums import ArtifactType
from aqueduct.error import AqueductError, InvalidArtifactTypeException


class GenericArtifact(BaseArtifact):
    """This class represents a generic artifact within the flow's DAG.

    Currently, a generic artifact can be any artifact other than table, numeric, bool, or parameter
    generated from eager execution, or an artifact of unknown type generated from lazy execution.
    """

    def __init__(
        self,
        dag: DAG,
        artifact_id: uuid.UUID,
        artifact_type: ArtifactType = ArtifactType.UNTYPED,
        content: Optional[Any] = None,
        from_flow_run: bool = False,
    ):
        self._dag = dag
        self._artifact_id = artifact_id
        # This parameter indicates whether the artifact is fetched from flow-run or not.
        self._from_flow_run = from_flow_run
        self._set_type(artifact_type)
        self._set_content(content)

        if self._from_flow_run:
            # If the artifact is initialized from a flow run, then it should not contain any content.
            assert self._get_content() is None

        if self._get_content() is not None:
            assert self._get_type() != ArtifactType.UNTYPED

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

        if parameters is None and self._get_content() is not None:
            return self._get_content()

        previewed_artifact = artifact_utils.preview_artifact(
            self._dag, self._artifact_id, parameters
        )
        if (
            self._get_type() != ArtifactType.UNTYPED
            and previewed_artifact._get_type() != self._get_type()
        ):
            raise InvalidArtifactTypeException(
                "The computed artifact is expected to be type %s, but has type %s"
                % (self._get_type(), previewed_artifact._get_type())
            )

        if parameters:
            return previewed_artifact._get_content()
        else:
            # We are materializing an artifact generated from lazy execution.
            assert self._get_content() is None
            self._set_content(previewed_artifact._get_content())
            self._set_type(previewed_artifact._get_type())
            return self._get_content()

    def describe(self) -> None:
        """Prints out a human-readable description of the bool artifact."""
        # TODO: make this more informative.
        print("This is a %s artifact." % self._get_type())
