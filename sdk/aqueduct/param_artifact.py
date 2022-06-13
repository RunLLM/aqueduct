import json
import uuid
from typing import Any, Dict, Optional

from aqueduct.api_client import APIClient
from aqueduct.dag import DAG
from aqueduct.error import InvalidUserArgumentException
from aqueduct.generic_artifact import Artifact


class ParamArtifact(Artifact):
    def __init__(
        self,
        api_client: APIClient,
        dag: DAG,
        artifact_id: uuid.UUID,
    ):
        """The APIClient is only included because decorated functions operators acting on this parameter
        will need a handle to an API client."""
        self._api_client = api_client
        self._dag = dag
        self._artifact_id = artifact_id

    def get(self, parameters: Optional[Dict[str, Any]] = None) -> Any:
        if parameters is not None:
            raise InvalidUserArgumentException(
                "Parameters cannot be supplied to parameter artifacts."
            )

        _ = self._dag.must_get_artifact(self._artifact_id)
        param_op = self._dag.must_get_operator(with_output_artifact_id=self._artifact_id)
        assert param_op.spec.param is not None, "Artifact is not a parameter."
        return json.loads(param_op.spec.param.val)

    def describe(self) -> None:
        print(
            f"""
            ==================== PARAMETER ARTIFACT =============================
            Name: '{self.name()}'
            Value: {self.get()}
            """
        )
