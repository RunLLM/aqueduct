import json
import uuid
from abc import ABC, abstractmethod
from typing import Any, Dict, Optional

from aqueduct.dag import DAG
from aqueduct.dag_deltas import (
    AddOrReplaceOperatorDelta,
    apply_deltas_to_dag,
    find_duplicate_load_operator,
)
from aqueduct.enums import ArtifactType, OperatorType
from aqueduct.error import (
    InvalidIntegrationException,
    InvalidUserActionException,
    InvalidUserArgumentException,
)
from aqueduct.operators import LoadSpec, Operator, OperatorSpec, S3LoadParams, SaveConfig
from aqueduct.utils import generate_uuid

from aqueduct import globals


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

    def _get_type(self) -> ArtifactType:
        return self._dag.must_get_artifact(artifact_id=self._artifact_id).type

    def _get_content(self) -> Any:
        return self._content

    def _set_content(self, content: Any) -> None:
        self._content = content

    def set_operator_type(self, operator_type: OperatorType) -> None:
        self._from_operator_type = operator_type

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

    def save(self, config: SaveConfig) -> None:
        """Configure this artifact to be written to a specific integration after it's computed in a published flow.

        Args:
            config:
                SaveConfig object generated from integration using
                the <integration>.config(...) method.
        Raises:
            InvalidIntegrationException:
                An error occurred because the requested integration could not be
                found.
            InvalidUserActionException:
                An error occurred because you are trying to load non-relational data into a relational integration.
            InvalidUserArgumentException:
                An error occurred because some necessary fields are missing in the SaveConfig.
        """
        integration_info = config.integration_info
        integration_load_params = config.parameters
        integrations_map = globals.__GLOBAL_API_CLIENT__.list_integrations()

        if integration_info.name not in integrations_map:
            raise InvalidIntegrationException("Not connected to db %s!" % integration_info.name)

        # Non-tabular data cannot be saved into relational data stores.
        if (
            self._get_type() not in [ArtifactType.UNTYPED, ArtifactType.TABLE]
            and integration_info.is_relational()
        ):
            raise InvalidUserActionException(
                "Unable to load non-relational data into relational data store `%s`."
                % integration_info.name
            )

        # Tabular data written into S3 must include a S3FileFormat hint.
        if self._get_type() == ArtifactType.TABLE and isinstance(config.parameters, S3LoadParams):
            if config.parameters.format is None:
                raise InvalidUserArgumentException(
                    "You must supply a file format when saving tabular data into S3 integration `%s`."
                    % integration_info.name
                )

        # We deduplicate load operators based on name (and therefore integration) AND
        # the input artifact. This allows multiple artifacts to write to the same integration,
        # as well as a single artifact to write to multiple integrations, all while keeping
        # the name of the load operator readable.
        load_op_name = "save to %s" % integration_info.name

        # Add the load operator as a terminal node.
        apply_deltas_to_dag(
            self._dag,
            deltas=[
                AddOrReplaceOperatorDelta(
                    op=Operator(
                        id=generate_uuid(),
                        name=load_op_name,
                        description="",
                        spec=OperatorSpec(
                            load=LoadSpec(
                                service=integration_info.service,
                                integration_id=integration_info.id,
                                parameters=integration_load_params,
                            )
                        ),
                        inputs=[self._artifact_id],
                    ),
                    output_artifacts=[],
                    find_duplicate_fn=find_duplicate_load_operator,
                ),
            ],
        )
