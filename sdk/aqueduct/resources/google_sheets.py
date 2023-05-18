from typing import Optional

from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.table_artifact import TableArtifact
from aqueduct.constants.enums import ArtifactType, GoogleSheetsSaveMode
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.operators import (
    ExtractSpec,
    GoogleSheetsExtractParams,
    GoogleSheetsLoadParams,
    Operator,
    OperatorSpec,
)
from aqueduct.models.resource import BaseResource, ResourceInfo
from aqueduct.resources.validation import validate_is_connected
from aqueduct.utils.dag_deltas import AddOperatorDelta, apply_deltas_to_dag
from aqueduct.utils.utils import generate_uuid

from ..utils.naming import default_artifact_name_from_op_name, sanitize_artifact_name
from .save import _save_artifact


class GoogleSheetsResource(BaseResource):
    """
    Class for Google Sheets resource.
    """

    def __init__(self, dag: DAG, metadata: ResourceInfo):
        self._dag = dag
        self._metadata = metadata

    @validate_is_connected()
    def spreadsheet(
        self,
        spreadsheet_id: str,
        name: Optional[str] = None,
        output: Optional[str] = None,
        description: str = "",
    ) -> TableArtifact:
        """
        Retrieves a spreadsheet from the Google Sheets resource.

        Args:
            spreadsheet_id:
                Id of spreadsheet to retrieve. This can be found in the URL of the spreadsheet, e.g.
                https://docs.google.com/spreadsheets/d/{SPREADSHEET_ID}/edit#gid=0
            name:
                Name of the query.
            output:
                Name to assign the output artifact. If not set, the default naming scheme will be used.
            description:
                Description of the query.

        Returns:
            TableArtifact representing the Google Sheet.
        """
        resource_info = self._metadata

        op_name = name or "%s query" % self.name()
        artifact_name = output or default_artifact_name_from_op_name(op_name)

        operator_id = generate_uuid()
        output_artifact_id = generate_uuid()
        apply_deltas_to_dag(
            self._dag,
            deltas=[
                AddOperatorDelta(
                    op=Operator(
                        id=operator_id,
                        name=op_name,
                        description=description,
                        spec=OperatorSpec(
                            extract=ExtractSpec(
                                service=resource_info.service,
                                resource_id=resource_info.id,
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
                            name=sanitize_artifact_name(artifact_name),
                            type=ArtifactType.TABLE,
                            explicitly_named=name is not None,
                        ),
                    ],
                )
            ],
        )

        return TableArtifact(
            dag=self._dag,
            artifact_id=output_artifact_id,
        )

    @validate_is_connected()
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
        _save_artifact(
            artifact.id(),
            self._dag,
            self._metadata,
            save_params=GoogleSheetsLoadParams(filepath=filepath, save_mode=save_mode),
        )

    def describe(self) -> None:
        """Prints out a human-readable description of the google sheets resource."""
        print("==================== Google Sheets Resource =============================")
        self._metadata.describe()
