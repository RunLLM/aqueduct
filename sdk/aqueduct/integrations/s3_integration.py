from typing import Optional

from aqueduct.api_client import APIClient
from aqueduct.artifact import Artifact, ArtifactSpec
from aqueduct.dag import DAG, apply_deltas_to_dag, AddOrReplaceOperatorDelta
from aqueduct.enums import S3FileFormat
from aqueduct.integrations.integration import IntegrationInfo, Integration
from aqueduct.operators import (
    Operator,
    OperatorSpec,
    ExtractSpec,
    S3ExtractParams,
    S3LoadParams,
    SaveConfig,
)
from aqueduct.table_artifact import TableArtifact
from aqueduct.utils import (
    generate_uuid,
    artifact_name_from_op_name,
    generate_extract_op_name,
)


class S3Integration(Integration):
    """
    Class for S3 integration.
    """

    def __init__(self, api_client: APIClient, dag: DAG, metadata: IntegrationInfo):
        self._api_client = api_client
        self._dag = dag
        self._metadata = metadata

    def file(
        self,
        filepath: str,
        format: S3FileFormat,
        name: Optional[str] = None,
        description: str = "",
    ) -> TableArtifact:
        """
        Retrieves a file from the S3 integration.

        Args:
            filepath:
                Filepath to retrieve from.
            name:
                Name of the query.
            description:
                Description of the query.

        Returns:
            TableArtifact representing the S3 File.
        """
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
                                parameters=S3ExtractParams(filepath=filepath, format=format),
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

        return TableArtifact(
            api_client=self._api_client,
            dag=self._dag,
            artifact_id=output_artifact_id,
        )

    def config(self, filepath: str, format: S3FileFormat) -> SaveConfig:
        """
        Configuration for saving to S3 Integration.

        Arguments:
            filepath:
                S3 Filepath to save to.
            format:
                S3 Fileformat to save as. Can be CSV, JSON, or Parquet.
        Returns:
            SaveConfig object to use in TableArtifact.save()
        """
        return SaveConfig(
            integration_info=self._metadata,
            parameters=S3LoadParams(filepath=filepath, format=format),
        )

    def describe(self) -> None:
        """Prints out a human-readable description of the S3 integration."""
        print("==================== S3 Integration  =============================")
        self._metadata.describe()
