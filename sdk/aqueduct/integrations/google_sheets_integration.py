from typing import Optional

from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.save import save_artifact
from aqueduct.artifacts.table_artifact import TableArtifact
from aqueduct.constants.enums import ArtifactType, GoogleSheetsSaveMode
from aqueduct.logger import logger
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.integration import Integration, IntegrationInfo
from aqueduct.models.operators import (
    ExtractSpec,
    GoogleSheetsExtractParams,
    GoogleSheetsLoadParams,
    Operator,
    OperatorSpec,
    SaveConfig,
)
from aqueduct.utils.dag_deltas import AddOrReplaceOperatorDelta, apply_deltas_to_dag
from aqueduct.utils.utils import artifact_name_from_op_name, generate_uuid

from .utils import _generate_extract_op_name


class GoogleSheetsIntegration(Integration):
    """
    Class for Google Sheets integration.
    """

    def __init__(self, dag: DAG, metadata: IntegrationInfo):
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
                                parameters=GoogleSheetsExtractParams(
                                    spreadsheet_id=spreadsheet_id,
                                ),
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

        return TableArtifact(
            dag=self._dag,
            artifact_id=output_artifact_id,
        )

    def config(
        self,
        filepath: str,
        save_mode: GoogleSheetsSaveMode = GoogleSheetsSaveMode.OVERWRITE,
    ) -> SaveConfig:
        """TODO(ENG-2035): Deprecated and will be removed.
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
        logger().warning(
            "`integration.config()` is deprecated. Please use `integration.save()` directly instead."
        )
        return SaveConfig(
            integration_info=self._metadata,
            parameters=GoogleSheetsLoadParams(filepath=filepath, save_mode=save_mode),
        )

    def save(
        self,
        artifact: BaseArtifact,
        filepath: str,
        save_mode: GoogleSheetsSaveMode = GoogleSheetsSaveMode.OVERWRITE,
    ) -> None:
        """Registers a save operator of the given artifact, to be executed when it's computed in a published flow.

        Args:
            artifact:
                The artifact to save into Google Sheets.
            filepath:
                The absolute file path to the Google Sheet to save to.
            save_mode:
                Defines the semantics of the save. Options are
                - "overwrite"
                - "create": Creates a new spreadsheet.
                            If the spreadsheet doesn't exist, has `overwrite` behavior.
                - "newsheet": Creates a new sheet in an existing spreadsheet.
                              If the spreadsheet doesn't exist, has `create` behavior.
        """
        save_artifact(
            artifact.id(),
            artifact.type(),
            self._dag,
            self._metadata,
            save_params=GoogleSheetsLoadParams(filepath=filepath, save_mode=save_mode),
        )

    def describe(self) -> None:
        """Prints out a human-readable description of the google sheets integration."""
        print("==================== Google Sheets Integration  =============================")
        self._metadata.describe()
