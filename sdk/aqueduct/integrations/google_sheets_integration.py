from typing import Optional

from aqueduct.api_client import APIClient
from aqueduct.artifact import Artifact, ArtifactSpec
from aqueduct.dag import DAG, apply_deltas_to_dag, AddOrReplaceOperatorDelta
from aqueduct.enums import GoogleSheetsSaveMode
from aqueduct.integrations.integration import IntegrationInfo, Integration
from aqueduct.operators import (
    Operator,
    OperatorSpec,
    ExtractSpec,
    GoogleSheetsExtractParams,
    GoogleSheetsLoadParams,
    SaveConfig,
)
from aqueduct.table_artifact import TableArtifact
from aqueduct.utils import (
    generate_uuid,
    artifact_name_from_op_name,
    generate_extract_op_name,
)


class GoogleSheetsIntegration(Integration):
    """
    Class for Google Sheets integration.
    """

    def __init__(self, api_client: APIClient, dag: DAG, metadata: IntegrationInfo):
        self._api_client = api_client
        self._dag = dag
        self._metadata = metadata

    def spreadsheet(
        self, spreadsheet_id: str, name: Optional[str] = None, description: str = ""
    ) -> TableArtifact:
        """
        Retrieves a spreadsheet from the Google Sheets integration.

        Args:
            spreadsheet_id:
                Id of spreadsheet to retrieve. This can be found in the URL of the spreadsheet, e.g.
                https://docs.google.com/spreadsheets/d/{SPREADSHEET_ID}/edit#gid=0
            name:
                Name of the query.
            description:
                Description of the query.

        Returns:
            TableArtifact representing the Google Sheet.
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
                                parameters=GoogleSheetsExtractParams(
                                    spreadsheet_id=spreadsheet_id,
                                ),
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

    def config(
        self,
        filepath: str,
        save_mode: GoogleSheetsSaveMode = GoogleSheetsSaveMode.OVERWRITE,
    ) -> SaveConfig:
        """
        Configuration for saving to Google Sheets Integration.

        Arguments:
            filepath:
                Google Sheets filepath to save to. Any directory in the path that does not exist
                will be created.
            save_mode:
                The save mode to use when saving this artifact to Google Sheets.
                Possible values are:
                - OVERWRITE: If a spreadsheet with the same name exists, it will overwrite it.
                    Otherwise, it will create a new spreadsheet.
                - NEWSHEET: If a spreadsheet with the same name exists, it will create a new sheet for this artifact.
                    Otherwise, it will create a new spreadsheet.
                - CREATE: This will create a new spreadsheet, even if one with the same name exists. The previous
                    spreadsheet will not be deleted, as Google Sheets allows for multiple spreadsheets of the same name.
        Returns:
            SaveConfig object to use in TableArtifact.save()
        """
        return SaveConfig(
            integration_info=self._metadata,
            parameters=GoogleSheetsLoadParams(filepath=filepath, save_mode=save_mode),
        )

    def describe(self) -> None:
        """Prints out a human-readable description of the google sheets integration."""
        print("==================== Google Sheets Integration  =============================")
        self._metadata.describe()
