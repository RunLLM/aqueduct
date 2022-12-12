import json
from typing import List, Optional, Union

from aqueduct.artifacts import preview as artifact_utils
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.save import save_artifact
from aqueduct.artifacts.transform import to_artifact_class
from aqueduct.constants.enums import ArtifactType, ExecutionMode, S3TableFormat
from aqueduct.logger import logger
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.integration import Integration, IntegrationInfo
from aqueduct.models.operators import (
    ExtractSpec,
    Operator,
    OperatorSpec,
    S3ExtractParams,
    S3LoadParams,
    SaveConfig,
)
from aqueduct.utils.dag_deltas import AddOrReplaceOperatorDelta, apply_deltas_to_dag
from aqueduct.utils.utils import artifact_name_from_op_name, generate_uuid

from aqueduct import globals

from .naming import _generate_extract_op_name


class S3Integration(Integration):
    """
    Class for S3 integration.
    """

    def __init__(self, dag: DAG, metadata: IntegrationInfo):
        self._dag = dag
        self._metadata = metadata

    def file(
        self,
        filepaths: Union[List[str], str],
        artifact_type: ArtifactType,
        format: Optional[str] = None,
        merge: Optional[bool] = None,
        name: Optional[str] = None,
        description: str = "",
        lazy: bool = False,
    ) -> BaseArtifact:
        """
        Reads one or more files from the S3 integration.

        Args:
            filepaths:
                Filepath to retrieve from. The filepaths can either be:
                1) a single string that represents a file name or a directory name. The directory
                name must ends with a `/`. In case of a file name, we attempt to retrieve that file,
                and in case of a directory name, we do a prefix search on the directory and retrieve
                all matched files and concatenate them into a single file.
                2) a list of strings representing the file name. Note that in this case, we do not
                accept directory names in the list.
            artifact_type:
                The expected type of the S3 files. The `ArtifactType` class in `enums.py` contains all
                supported types, except for ArtifactType.UNTYPED. Note that when multiple files are
                retrieved, they must have the same artifact type.
            format:
                If the artifact type is ArtifactType.TABLE, the user has to specify the table format.
                We currently support JSON, CSV, and Parquet. Note that when multiple files are retrieved,
                they must have the same format.
            merge:
                If the artifact type is ArtifactType.TABLE, we can optionally merge multiple tables
                into a single DataFrame if this flag is set to True.
            name:
                Name of the query.
            description:
                Description of the query.

        Returns:
            Artifact or a tuple of artifacts representing the S3 Files.
        """
        if globals.__GLOBAL_CONFIG__.lazy:
            lazy = True
        execution_mode = ExecutionMode.EAGER if not lazy else ExecutionMode.LAZY

        if format:
            if artifact_type != ArtifactType.TABLE:
                raise Exception(
                    "Format argument is only applicable to table artifact type, found %s instead."
                    % artifact_type
                )

            lowercased_format = format.lower()
            if lowercased_format == S3TableFormat.CSV.value.lower():
                format_enum = S3TableFormat.CSV
            elif lowercased_format == S3TableFormat.JSON.value.lower():
                format_enum = S3TableFormat.JSON
            elif lowercased_format == S3TableFormat.PARQUET.value.lower():
                format_enum = S3TableFormat.PARQUET
            else:
                raise Exception("Unsupport file format %s." % format)
        else:
            format_enum = None

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
                                parameters=S3ExtractParams(
                                    filepath=json.dumps(filepaths),
                                    artifact_type=artifact_type,
                                    format=format_enum,
                                    merge=merge,
                                ),
                            )
                        ),
                        outputs=[output_artifact_id],
                    ),
                    output_artifacts=[
                        ArtifactMetadata(
                            id=output_artifact_id,
                            name=artifact_name_from_op_name(op_name),
                            type=artifact_type,
                        ),
                    ],
                )
            ],
        )

        if execution_mode == ExecutionMode.EAGER:
            # Issue preview request since this is an eager execution.
            return artifact_utils.preview_artifact(self._dag, output_artifact_id)
        else:
            # We are in lazy mode.
            return to_artifact_class(self._dag, output_artifact_id, artifact_type)

    def config(self, filepath: str, format: Optional[S3TableFormat] = None) -> SaveConfig:
        """TODO(ENG-2035): Deprecated and will be removed.
        Configuration for saving to S3 Integration.

        Arguments:
            filepath:
                S3 Filepath to save to.
            format:
                S3 Fileformat to save as. Can be CSV, JSON, or Parquet.
        Returns:
            SaveConfig object to use in Artifact.save()
        """
        logger().warning(
            "`integration.config()` is deprecated. Please use `integration.save()` directly instead."
        )
        return SaveConfig(
            integration_info=self._metadata,
            parameters=S3LoadParams(filepath=filepath, format=format),
        )

    def save(
        self, artifact: BaseArtifact, filepath: str, format: Optional[S3TableFormat] = None
    ) -> None:
        """Registers a save operator of the given artifact, to be executed when it's computed in a published flow.

        Args:
            artifact:
                The artifact to save into S3.
            filepath:
                The S3 path to save to. Will overwrite any existing object at that path.
            format:
                Defines the format that the artifact will be saved as.
                Options are "CSV", "JSON", "Parquet".
        """
        save_artifact(
            artifact.id(),
            artifact.type(),
            self._dag,
            self._metadata,
            save_params=S3LoadParams(filepath=filepath, format=format),
        )

    def describe(self) -> None:
        """Prints out a human-readable description of the S3 integration."""
        print("==================== S3 Integration  =============================")
        self._metadata.describe()
