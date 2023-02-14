import json
import uuid
from abc import ABC, abstractmethod
from typing import Any, Dict, Optional

from aqueduct.constants.enums import ArtifactType, OperatorType
from aqueduct.error import InvalidUserActionException
from aqueduct.models.dag import DAG
from aqueduct.models.operators import get_operator_type


class BaseArtifact(ABC):
    _artifact_id: uuid.UUID
    _dag: DAG
    _content: Any
    _from_flow_run: bool
    _from_operator_type: Optional[OperatorType] = None

    def id(self) -> uuid.UUID:
        """Fetch the id associated with this artifact.

        This id will not exist in the system if the artifact has not yet been published.
        """
        return self._artifact_id

    def name(self) -> str:
        """Fetch the name of this artifact."""
        return self._dag.must_get_artifact(artifact_id=self._artifact_id).name

    def type(self) -> ArtifactType:
        return self._dag.must_get_artifact(artifact_id=self._artifact_id).type

    def _get_content(self) -> Any:
        return self._content

    def _set_content(self, content: Any) -> None:
        self._content = content

    def set_operator_type(self, operator_type: OperatorType) -> None:
        self._from_operator_type = operator_type

    def set_name(self, name: str) -> None:
        self._dag.validate_artifact_name(name)

        # If this a parameter artifact, we will also need to change the name of the parameter,
        # to preserve our invariant that a param op and its artifact always have the same name.
        op = self._dag.must_get_operator(with_output_artifact_id=self._artifact_id)
        if get_operator_type(op) == OperatorType.PARAM:
            if self._dag.get_operator(with_name=name) is not None:
                raise InvalidUserActionException(
                    "Unable to change parameter name to %s, there already exists an operator with the same name. "
                    "Parameter names must be globally unique." % name,
                )
            self._dag.update_operator_name(op.id, name)

        # Update the name of the artifact.
        self._dag.update_artifact_name(self._artifact_id, name)

    def _describe(self) -> Dict[str, Any]:
        input_operator = self._dag.must_get_operator(with_output_artifact_id=self._artifact_id)
        return {
            "Id": str(self._artifact_id),
            "Label": input_operator.name,
            "Spec": json.loads(input_operator.spec.json(exclude_none=True)),
        }

    @abstractmethod
    def describe(self) -> None:
        pass

    @abstractmethod
    def get(self, parameters: Optional[Dict[str, Any]] = None) -> Any:
        pass
