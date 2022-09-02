from __future__ import annotations

import json
import uuid
from typing import Any, Dict, Optional, Union

import numpy as np
from aqueduct.artifacts import utils as artifact_utils
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.dag import DAG
from aqueduct.enums import ArtifactType
from aqueduct.error import AqueductError, InvalidArtifactTypeException
from aqueduct.utils import format_header_for_print, get_description_for_check


class BoolArtifact(BaseArtifact):
    """This class represents a bool within the flow's DAG.

    Any `@check`-annotated python function that returns a boolean will
    return this class when that function is called. This class can also
    be returned from pre-defined functions like metric.bound(...).

    Any `@op`-annotated python function that returns a boolean
    will return this class when that function is called in non-lazy mode.

    Examples:
        >>> @check
        >>> def check_something(df1, metric) -> bool:
        >>>     return check_result
        >>>
        >>> check_artifact = check_something(table_artifact, metric_artifact)

        The contents of the bool artifact can be manifested locally:

        >>> assert check_artifact.get()
    """

    def __init__(
        self,
        dag: DAG,
        artifact_id: uuid.UUID,
        content: Optional[Union[bool, np.bool_]] = None,
        from_flow_run: bool = False,
    ):
        self._dag = dag
        self._artifact_id = artifact_id
        # This parameter indicates whether the artifact is fetched from flow-run or not.
        self._from_flow_run = from_flow_run
        self._set_content(content)
        if self._from_flow_run:
            # If the artifact is initialized from a flow run, then it should not contain any content.
            assert self._get_content() is None

        self._set_type(ArtifactType.BOOL)

    def get(self, parameters: Optional[Dict[str, Any]] = None) -> Union[bool, np.bool_]:
        """Materializes a BoolArtifact into a boolean.

        Returns:
            A boolean representing whether the check passed or not.

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
        if previewed_artifact._get_type() != ArtifactType.BOOL:
            raise InvalidArtifactTypeException(
                "Error: the computed result is expected to of type bool, found %s"
                % previewed_artifact._get_type()
            )

        assert isinstance(previewed_artifact._get_content(), bool) or isinstance(
            previewed_artifact._get_content(), np.bool_
        )

        if parameters:
            return previewed_artifact._get_content()
        else:
            # We are materializing an artifact generated from lazy execution.
            assert self._get_content() is None
            self._set_content(previewed_artifact._get_content())
            return self._get_content()

    def describe(self) -> None:
        """Prints out a human-readable description of the bool artifact."""
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

        print(format_header_for_print(f"'{input_operator.name}' Bool Artifact"))
        print(json.dumps(readable_dict, sort_keys=False, indent=4))
