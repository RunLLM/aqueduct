import uuid
from typing import Optional

from aqueduct.api_client import APIClient
from aqueduct.artifact import Artifact, ArtifactSpec
from aqueduct.dag import DAG, apply_deltas_to_dag, AddOrReplaceOperatorDelta
from aqueduct.enums import SalesforceExtractType
from aqueduct.integrations.integration import IntegrationInfo, Integration
from aqueduct.operators import (
    Operator,
    OperatorSpec,
    ExtractSpec,
    SalesforceExtractParams,
    SalesforceLoadParams,
    SaveConfig,
)
from aqueduct.table_artifact import TableArtifact
from aqueduct.utils import (
    generate_uuid,
    artifact_name_from_op_name,
    generate_extract_op_name,
)


class SalesforceIntegration(Integration):
    """
    Class for Salesforce integration.
    """

    def __init__(self, api_client: APIClient, dag: DAG, metadata: IntegrationInfo):
        self._api_client = api_client
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
            api_client=self._api_client,
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
            api_client=self._api_client,
            dag=self._dag,
            artifact_id=output_artifact_id,
        )

    def config(self, object: str) -> SaveConfig:
        """
        Configuration for saving to Salesforce Integration.

        Arguments:
            object:
                Object to save to.
        Returns:
            SaveConfig object to use in TableArtifact.save()
        """
        return SaveConfig(
            integration_info=self._metadata,
            parameters=SalesforceLoadParams(object=object),
        )

    def _add_extract_operation(
        self,
        name: Optional[str],
        description: str,
        query: str,
        extract_type: SalesforceExtractType,
    ) -> uuid.UUID:

        integration_info = self._metadata

        op_name = generate_extract_op_name(self._dag, integration_info.name, name)

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
                        Artifact(
                            id=output_artifact_id,
                            name=artifact_name_from_op_name(op_name),
                            spec=ArtifactSpec(table={}),
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
