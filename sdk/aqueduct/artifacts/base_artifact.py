import json
import uuid
from abc import ABC, abstractmethod
from typing import Any, Dict, Optional, Union

import numpy as np
from aqueduct.constants.enums import ArtifactType, OperatorType
from aqueduct.models.dag import DAG
from aqueduct.models.execution_state import ExecutionState
from aqueduct.type_annotations import Number
from aqueduct.utils.naming import sanitize_artifact_name


class BaseArtifact(ABC):
    _artifact_id: uuid.UUID
    _dag: DAG
    _content: Any
    _from_flow_run: bool
    _from_operator_type: Optional[OperatorType] = None
    _execution_state: Optional[ExecutionState] = None

    def __init__(
        self,
        dag: DAG,
        artifact_id: uuid.UUID,
        content: Optional[Union[bool, np.bool_, Number]] = None,
        from_flow_run: bool = False,
        execution_state: Optional[ExecutionState] = None,
    ) -> None:
        self._dag = dag
        self._artifact_id = artifact_id

        # This parameter indicates whether the artifact is fetched from flow-run or not.
        self._from_flow_run = from_flow_run
        self._set_content(content)

        # For now, the execution_state is only relevant when it's fetched from a flow run.
        # It stays 'None' when the artifact runs in previews.
        self._execution_state = execution_state

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

    def snapshot_enabled(self) -> bool:
        return self._dag.must_get_artifact(artifact_id=self._artifact_id).should_persist

    def execution_state(self) -> Optional[ExecutionState]:
        return self._execution_state

    def _get_content(self) -> Any:
        return self._content

    def _set_content(self, content: Any) -> None:
        self._content = content

    def set_operator_type(self, operator_type: OperatorType) -> None:
        self._from_operator_type = operator_type

    def set_name(self, name: str) -> None:
        self._dag.update_artifact_name(self._artifact_id, sanitize_artifact_name(name))

    def enable_snapshot(self) -> None:
        self._dag.update_artifact_should_persist(self._artifact_id, True)

    def disable_snapshot(self) -> None:
        self._dag.update_artifact_should_persist(self._artifact_id, False)

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
