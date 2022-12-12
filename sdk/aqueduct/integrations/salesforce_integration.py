import uuid
from typing import Optional

from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.save import save_artifact
from aqueduct.artifacts.table_artifact import TableArtifact
from aqueduct.constants.enums import ArtifactType, SalesforceExtractType
from aqueduct.logger import logger
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.integration import Integration, IntegrationInfo
from aqueduct.models.operators import (
    ExtractSpec,
    Operator,
    OperatorSpec,
    SalesforceExtractParams,
    SalesforceLoadParams,
    SaveConfig,
)
from aqueduct.utils.dag_deltas import AddOrReplaceOperatorDelta, apply_deltas_to_dag
from aqueduct.utils.utils import artifact_name_from_op_name, generate_uuid

from .naming import _generate_extract_op_name


class SalesforceIntegration(Integration):
    """
    Class for Salesforce integration.
    """

    def __init__(self, dag: DAG, metadata: IntegrationInfo):
        self._dag = dag
        self._metadata = metadata

    def search(
        self, search_query: str, name: Optional[str] = None, description: str = ""
    ) -> TableArtifact:
        """
        Runs a search against the Salesforce integration.

        Args:
            search_query:
                The search query to run.
            name:
                Name of the query.
            description:
                Description of the query.

        Returns:
            TableArtifact representing result of the SQL query.
        """
        output_artifact_id = self._add_extract_operation(
            name, description, search_query, SalesforceExtractType.SEARCH
        )

        return TableArtifact(
            dag=self._dag,
            artifact_id=output_artifact_id,
        )

    def query(self, query: str, name: Optional[str] = None, description: str = "") -> TableArtifact:
        """
        Runs a query against the Salesforce integration.

        Args:
            query:
                The query to run.
            name:
                Name of the query.
            description:
                Description of the query.

        Returns:
            TableArtifact representing result of the SQL query.
        """
        output_artifact_id = self._add_extract_operation(
            name, description, query, SalesforceExtractType.QUERY
        )

        return TableArtifact(
            dag=self._dag,
            artifact_id=output_artifact_id,
        )

    def config(self, object: str) -> SaveConfig:
        """TODO(ENG-2035): Deprecated and will be removed.
        Configuration for saving to Salesforce Integration.

        Arguments:
            object:
                Object to save to.
        Returns:
            SaveConfig object to use in TableArtifact.save()
        """
        logger().warning(
            "`integration.config()` is deprecated. Please use `integration.save()` directly instead."
        )
        return SaveConfig(
            integration_info=self._metadata,
            parameters=SalesforceLoadParams(object=object),
        )

    def save(self, artifact: BaseArtifact, object: str) -> None:
        """Registers a save operator of the given artifact, to be executed when it's computed in a published flow.

        Args:
            artifact:
                The artifact to save into Salesforce.
            object:
                The name of the Salesforce object to save to.
        """
        save_artifact(
            artifact.id(),
            artifact.type(),
            self._dag,
            self._metadata,
            save_params=SalesforceLoadParams(object=object),
        )

    def _add_extract_operation(
        self,
        name: Optional[str],
        description: str,
        query: str,
        extract_type: SalesforceExtractType,
    ) -> uuid.UUID:

        integration_info = self._metadata

        op_name = _generate_extract_op_name(self._dag, integration_info.name, name)

        operator_id = generate_uuid()
        output_artifact_id = generate_uuid()
        apply_deltas_to_dag(
            self._dag,
            deltas=[
                AddOrReplaceOperatorDelta(
                    op=Operator(
                        id=operator_id,
                        name=op_name,
                        description=description,
                        spec=OperatorSpec(
                            extract=ExtractSpec(
                                service=integration_info.service,
                                integration_id=integration_info.id,
                                parameters=SalesforceExtractParams(type=extract_type, query=query),
                            )
                        ),
                        outputs=[output_artifact_id],
                    ),
                    output_artifacts=[
                        ArtifactMetadata(
                            id=output_artifact_id,
                            name=artifact_name_from_op_name(op_name),
                            type=ArtifactType.TABLE,
                        ),
                    ],
                )
            ],
        )

        return output_artifact_id

    def describe(self) -> None:
        """Prints out a human-readable description of the Salesforce integration."""
        print("==================== Salesforce Integration  =============================")
        self._metadata.describe()
