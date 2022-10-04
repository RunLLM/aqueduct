import json
from typing import List, Optional, Union

from aqueduct.artifacts import utils as artifact_utils
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.metadata import ArtifactMetadata
from aqueduct.artifacts.utils import to_artifact_class
from aqueduct.dag import DAG
from aqueduct.dag_deltas import AddOrReplaceOperatorDelta, apply_deltas_to_dag
from aqueduct.enums import ArtifactType, ExecutionMode, S3TableFormat
from aqueduct.integrations.integration import Integration, IntegrationInfo
from aqueduct.operators import (
    ExtractSpec,
    Operator,
    OperatorSpec,
    S3ExtractParams,
    S3LoadParams,
    SaveConfig,
)
from aqueduct.utils import artifact_name_from_op_name, generate_extract_op_name, generate_uuid

from aqueduct import globals


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
        """
        Configuration for saving to S3 Integration.

        Arguments:
            filepath:
                S3 Filepath to save to.
            format:
                S3 Fileformat to save as. Can be CSV, JSON, or Parquet.
        Returns:
            SaveConfig object to use in Artifact.save()
        """
        return SaveConfig(
            integration_info=self._metadata,
            parameters=S3LoadParams(filepath=filepath, format=format),
        )

    def describe(self) -> None:
        """Prints out a human-readable description of the S3 integration."""
        print("==================== S3 Integration  =============================")
        self._metadata.describe()
