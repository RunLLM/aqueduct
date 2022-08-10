from __future__ import annotations

import json
import uuid
from typing import Any, Dict, Optional

from aqueduct.dag import DAG, SubgraphDAGDelta, UpdateParametersDelta, apply_deltas_to_dag
from aqueduct.error import AqueductError
from aqueduct.generic_artifact import Artifact
from aqueduct.utils import format_header_for_print, get_description_for_check

from aqueduct import api_client
from aqueduct.enums import ArtifactType


class BoolArtifact(Artifact):
    """This class represents a check within the flow's DAG.

    Any `@check`-annotated python function that returns a boolean will
    return this class when that function is called called. This also
    is returned from pre-defined functions like metric.bound(...).

    Examples:
        >>> @check
        >>> def check_something(df1, metric) -> bool:
        >>> return check_result
        >>>
        >>> check_artifact = check_something(table_artifact, metric_artifact)

        The contents of the check artifact can be manifested locally:

        >>> assert check_artifact.get()
    """

    def __init__(self, dag: DAG, artifact_id: uuid.UUID, content: bool, from_flow_run: bool = False):
        self._dag = dag
        self._artifact_id = artifact_id
        # This parameter indicates whether the artifact is fetched from flow-run or not.
        self._from_flow_run = from_flow_run
        self._content = content
        self._type = ArtifactType.BOOL

    def get(self, parameters: Optional[Dict[str, Any]] = None) -> bool:
        """Materializes a CheckArtifact into a boolean.

        Returns:
            A boolean representing whether the check passed or not.

        Raises:
            InvalidRequestError:
                An error occurred because of an issue with the user's code or inputs.
            InternalServerError:
                An unexpected error occurred in the server.
        """
        if parameters:
            artifact_response = preview_artifact(self._dag, self._artifact_id, parameters)
            if artifact.type() != ArtifactType.BOOL:
                raise Exception("Error: the computed result is expected to of type bool, found %s" % artifact.type())
            return artifact._content()
        else:
            return self._content

    def describe(self) -> None:
        """Prints out a human-readable description of the check artifact."""
        input_operator = self._dag.must_get_operator(with_output_artifact_id=self._artifact_id)

        general_dict = get_description_for_check(input_operator)

        # Remove because values already in `readable_dict`
        general_dict.pop("Label")
        general_dict.pop("Level")

        readable_dict = super()._describe()
        readable_dict.update(general_dict)
        readable_dict["Inputs"] = [
            self._dag.must_get_artifact(artf).name for artf in input_operator.inputs
        ]

        print(format_header_for_print(f"'{input_operator.name}' Check Artifact"))
        print(json.dumps(readable_dict, sort_keys=False, indent=4))
