import json
from typing import List, Optional, Union

from aqueduct.artifacts import preview as artifact_utils
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.constants.enums import ArtifactType, ExecutionMode, S3TableFormat
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.integration import Integration, IntegrationInfo
from aqueduct.models.operators import (
    ExtractSpec,
    Operator,
    OperatorSpec,
    S3ExtractParams,
    S3LoadParams,
)
from aqueduct.utils.dag_deltas import AddOperatorDelta, apply_deltas_to_dag
from aqueduct.utils.utils import generate_uuid

from aqueduct import globals

from ..artifacts.create import to_artifact_class
from ..error import InvalidUserArgumentException
from ..utils.naming import default_artifact_name_from_op_name, sanitize_artifact_name
from .save import _save_artifact


def _convert_to_s3_table_format(format: Optional[str]) -> Optional[S3TableFormat]:
    """A simple string -> enum conversion. Returns None if no format provided."""
    if format is None:
        return None

    lowercased_format = format.lower()
    if lowercased_format == S3TableFormat.CSV.value.lower():
        format_enum = S3TableFormat.CSV
    elif lowercased_format == S3TableFormat.JSON.value.lower():
        format_enum = S3TableFormat.JSON
    elif lowercased_format == S3TableFormat.PARQUET.value.lower():
        format_enum = S3TableFormat.PARQUET
    else:
        raise InvalidUserArgumentException("Unsupported S3 file format `%s`." % format)
    return format_enum


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
        output: Optional[str] = None,
        description: str = "",
        lazy: bool = False,
    ) -> BaseArtifact:
        """
        Reads one or more files from the S3 integration.

        Args:
            filepaths:
                Filepath to retrieve from. The filepaths can either be:
                1) a single string that represents a file name or a directory name. The directory
                name must ends with a `/`. In case of a file name, we attempt to retrieve that file.
                In case of a directory name, we do a prefix search on the directory and retrieve
                all matched files in alphabetical order, returning them as a TUPLE artifact.
                2) a list of strings representing the file name. Note that in this case, we do not
                accept directory names in the list. The fetched data in this case will always be of
                ArtifactType.TUPLE.
            artifact_type:
                The expected type of the S3 files. The `ArtifactType` class in `enums.py` contains all
                supported types, except for ArtifactType.UNTYPED. Note that when multiple files are
                retrieved, they must have the same artifact type.
            format:
                If the artifact type is ArtifactType.TABLE, the user has to specify the table format.
                We currently support JSON, CSV, and Parquet. Note that when multiple table files are
                retrieved, they must have the same format.
            merge:
                If the artifact type is ArtifactType.TABLE, we can optionally merge multiple tables
                into a single DataFrame if this flag is set to True. This merge is done with
                `pandas.concat(tables, ignore_index=True)`.
            name:
                Name of the query.
            output:
                Name to assign the output artifact. If not set, the default naming scheme will be used.
            description:
                Description of the query.
            lazy:
                Whether to run this operator lazily. See https://docs.aqueducthq.com/operators/lazy-vs.-eager-execution .

        Returns:
            An artifact representing the S3 File(s). If multiple files are expected, the artifact
            will represent a tuple.
        """
        if globals.__GLOBAL_CONFIG__.lazy:
            lazy = True
        execution_mode = ExecutionMode.EAGER if not lazy else ExecutionMode.LAZY

        if format and artifact_type != ArtifactType.TABLE:
            raise InvalidUserArgumentException(
                "Format argument is only applicable to table artifact type, found %s instead."
                % artifact_type
            )
        format_enum = _convert_to_s3_table_format(format)

        integration_info = self._metadata
        op_name = name or "%s query" % self.name()
        artifact_name = output or default_artifact_name_from_op_name(op_name)

        operator_id = generate_uuid()
        output_artifact_id = generate_uuid()

        def _is_directory_search() -> bool:
            return isinstance(filepaths, str) and filepaths[-1] == "/"

        def _is_multi_file_search() -> bool:
            return isinstance(filepaths, list)

        # We expect a tuple output if multiple files are being fetched (unmerged), either due to
        # multi-file or directory search.
        output_artifact_type = artifact_type
        if not merge and (_is_directory_search() or _is_multi_file_search()):
            output_artifact_type = ArtifactType.TUPLE

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
                            name=sanitize_artifact_name(artifact_name),
                            type=output_artifact_type,
                            explicitly_named=output is not None,
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

    def save(self, artifact: BaseArtifact, filepath: str, format: Optional[str] = None) -> None:
        """Registers a save operator of the given artifact, to be executed when it's computed in a published flow.

        Args:
            artifact:
                The artifact to save into S3.
            filepath:
                The S3 path to save to. Will overwrite any existing object at that path.
            format:
                Only required if saving a table artifact. Options are case-insensitive "json", "csv", "parquet".
        """
        if artifact.type() == ArtifactType.TABLE and format is None:
            raise InvalidUserArgumentException(
                "You must supply a file format when saving tabular data into S3 integration `%s`."
                % self.name(),
            )
        elif (
            artifact.type() != ArtifactType.TABLE
            and artifact.type() != ArtifactType.UNTYPED
            and format is not None
        ):
            raise InvalidUserArgumentException(
                "A `format` argument should only be supplied for saving table artifacts. This artifact type is %s."
                % artifact.type()
            )

        _save_artifact(
            artifact.id(),
            self._dag,
            self._metadata,
            save_params=S3LoadParams(filepath=filepath, format=_convert_to_s3_table_format(format)),
        )

    def describe(self) -> None:
        """Prints out a human-readable description of the S3 integration."""
        print("==================== S3 Integration  =============================")
        self._metadata.describe()
