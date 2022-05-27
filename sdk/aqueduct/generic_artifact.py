from abc import ABC, abstractmethod
import json
from typing import Any, Dict
import uuid
from aqueduct.dag import DAG


class Artifact(ABC):

    _artifact_id: uuid.UUID
    _dag: DAG

    def id(self) -> uuid.UUID:
        """Fetch the id associated with this artifact.

        This id will not exist in the system if the artifact has not yet been published.
        """
        return self._artifact_id

    def name(self) -> str:
        """Fetch the name of this artifact."""
        return self._dag.must_get_artifact(artifact_id=self._artifact_id).name

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
    def get(self) -> Any:
        pass
