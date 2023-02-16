import uuid
from typing import Optional

from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.table_artifact import TableArtifact
from aqueduct.constants.enums import ArtifactType, SalesforceExtractType
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.integration import Integration, IntegrationInfo
from aqueduct.models.operators import (
    ExtractSpec,
    Operator,
    OperatorSpec,
    SalesforceExtractParams,
    SalesforceLoadParams,
)
from aqueduct.utils.dag_deltas import AddOrReplaceOperatorDelta, apply_deltas_to_dag
from aqueduct.utils.utils import generate_uuid

from .naming import _resolve_op_and_artifact_name_for_extract
from .save import _save_artifact


class SalesforceIntegration(Integration):
    """
    Class for Salesforce integration.
    """

    def __init__(self, dag: DAG, metadata: IntegrationInfo):
        self._dag = dag
        self._metadata = metadata

    def search(
        self,
        search_query: str,
        name: Optional[str] = None,
        output: Optional[str] = None,
        description: str = "",
    ) -> TableArtifact:
        """
        Runs a search against the Salesforce integration.

        Args:
            search_query:
                The search query to run.
            name:
                Name of the query.
           output:
                Name to assign the output artifact. If not set, the default naming scheme will be used.
            description:
                Description of the query.

        Returns:
            TableArtifact representing result of the SQL query.
        """

        op_name, artifact_name = _resolve_op_and_artifact_name_for_extract(
            dag=self._dag,
            op_name=name,
            default_op_name="%s search" % self.name(),
            artifact_name=output,
        )
        output_artifact_id = self._add_extract_operation(
            op_name, artifact_name, description, search_query, SalesforceExtractType.SEARCH
        )

        return TableArtifact(
            dag=self._dag,
            artifact_id=output_artifact_id,
        )

    def query(
        self,
        query: str,
        name: Optional[str] = None,
        output: Optional[str] = None,
        description: str = "",
    ) -> TableArtifact:
        """
        Runs a query against the Salesforce integration.

        Args:
            query:
                The query to run.
            name:
                Name of the query.
            output:
                Name to assign the output artifact. If not set, the default naming scheme will be used.
            description:
                Description of the query.

        Returns:
            TableArtifact representing result of the SQL query.
        """

        op_name, artifact_name = _resolve_op_and_artifact_name_for_extract(
            dag=self._dag,
            op_name=name,
            default_op_name="%s query" % self.name(),
            artifact_name=output,
        )

        output_artifact_id = self._add_extract_operation(
            op_name, artifact_name, description, query, SalesforceExtractType.QUERY
        )

        return TableArtifact(
            dag=self._dag,
            artifact_id=output_artifact_id,
        )

    def save(self, artifact: BaseArtifact, object: str) -> None:
        """Registers a save operator of the given artifact, to be executed when it's computed in a published flow.

        Args:
            artifact:
                The artifact to save into Salesforce.
            object:
                The name of the Salesforce object to save to.
        """
        _save_artifact(
            artifact.id(),
            self._dag,
            self._metadata,
            save_params=SalesforceLoadParams(object=object),
        )

    def _add_extract_operation(
        self,
        name: Optional[str],
        artifact_name: str,
        description: str,
        query: str,
        extract_type: SalesforceExtractType,
    ) -> uuid.UUID:
        integration_info = self._metadata

        operator_id = generate_uuid()
        output_artifact_id = generate_uuid()
        apply_deltas_to_dag(
            self._dag,
            deltas=[
                AddOrReplaceOperatorDelta(
                    op=Operator(
                        id=operator_id,
                        name=name,
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
                            name=artifact_name,
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
